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

func CryptoMiddleware(keyPath string) func(http.Handler) http.Handler {
	var privateKey *rsa.PrivateKey

	// If keyPath is nil, skip loading key entirely
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
			if keyPath == "" {
				next.ServeHTTP(w, r)
				return
			}

			encBody, err := io.ReadAll(r.Body)
			r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("failed to read request body"))
				return
			}

			if len(encBody) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			cipherText, err := base64.StdEncoding.DecodeString(string(encBody))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("invalid base64 body"))
				return
			}

			plainText, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, cipherText)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("failed to decrypt body"))
				return
			}

			r.Body = io.NopCloser(strings.NewReader(string(plainText)))
			r.ContentLength = int64(len(plainText))

			next.ServeHTTP(w, r)
		})
	}
}
