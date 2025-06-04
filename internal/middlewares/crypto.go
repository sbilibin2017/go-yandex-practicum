package middlewares

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"go.uber.org/zap"
)

// CryptoMiddleware возвращает HTTP middleware, который при наличии пути к приватному ключу
// загружает RSA приватный ключ и расшифровывает тело входящих запросов.
//
// Параметры:
//   - keyPath: путь к файлу с приватным RSA ключом в PEM формате. Если пустая строка,
//     расшифровка не выполняется и тело запроса передается как есть.
//
// Поведение middleware:
//   - Считывает тело запроса полностью.
//   - Если приватный ключ загружен, пытается декодировать тело из base64,
//     затем расшифровывает с помощью RSA PKCS1v15.
//   - Заменяет тело запроса на расшифрованное (или исходное, если ключ отсутствует).
//   - Устанавливает корректный ContentLength.
//
// Использование:
//
//	Этот middleware полезен для серверов, которые принимают зашифрованные запросы
//	с помощью публичного ключа клиента и хотят автоматически расшифровать их.
//
// Возвращает функцию middleware для использования в цепочке HTTP-обработчиков.
func CryptoMiddleware(keyPath string) func(http.Handler) http.Handler {
	var privateKey *rsa.PrivateKey

	if keyPath != "" {
		keyData, err := os.ReadFile(keyPath)
		if err != nil {
			logger.Log.Error("CryptoMiddleware: failed to read private key file", zap.Error(err))
		} else {
			block, _ := pem.Decode(keyData)
			if block == nil || block.Type != "RSA PRIVATE KEY" {
				logger.Log.Error("CryptoMiddleware: failed to decode PEM block containing private key")
			} else {
				privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
				if err != nil {
					logger.Log.Error("CryptoMiddleware: failed to parse private key", zap.Error(err))
				}
			}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			encBody, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read request body", http.StatusBadRequest)
				return
			}
			r.Body.Close()

			if len(encBody) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			var plainText []byte
			if privateKey != nil {
				cipherText, err := base64.StdEncoding.DecodeString(string(encBody))
				if err != nil {
					http.Error(w, "invalid base64 body", http.StatusBadRequest)
					return
				}

				plainText, err = rsa.DecryptPKCS1v15(rand.Reader, privateKey, cipherText)
				if err != nil {
					http.Error(w, "failed to decrypt body", http.StatusBadRequest)
					return
				}
			} else {
				plainText = encBody
			}

			r.Body = io.NopCloser(strings.NewReader(string(plainText)))
			r.ContentLength = int64(len(plainText))

			next.ServeHTTP(w, r)
		})
	}
}
