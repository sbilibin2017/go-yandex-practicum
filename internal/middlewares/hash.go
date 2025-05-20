package middlewares

import (
	"bytes"
	"io"
	"net/http"
)

// HashMiddleware проверяет и добавляет хеш-подпись (например, HMAC) к HTTP-запросам и ответам.
//
// Он выполняет следующие действия:
//   - Считывает тело запроса и вычисляет его хеш с помощью предоставленной hashFunc.
//   - Сравнивает вычисленный хеш с тем, что передан в заголовке запроса (header).
//   - Если хеши не совпадают, возвращает HTTP 400 (hash mismatch).
//   - Если хеши совпадают или не переданы, обрабатывает запрос.
//   - После выполнения обработчика вычисляет хеш для тела ответа и устанавливает его в заголовок.
//
// Аргументы:
//   - key: секретный ключ для вычисления хеша (например, HMAC).
//   - header: имя заголовка, в котором ожидается и устанавливается хеш.
//   - hashFunc: функция, вычисляющая хеш от данных и ключа.
//   - compareFunc: функция, безопасно сравнивающая два хеша (например, hmac.Equal).
func HashMiddleware(
	key string,
	header string,
	hashFunc func(data []byte, key string) string,
	compareFunc func(hash1 string, hash2 string) bool,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if key == "" {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Чтение тела запроса
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read request body", http.StatusInternalServerError)
				return
			}
			r.Body.Close()
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// Проверка подписи запроса
			receivedHash := r.Header.Get(header)
			if receivedHash != "" {
				computedHash := hashFunc(bodyBytes, key)
				if !compareFunc(receivedHash, computedHash) {
					http.Error(w, "hash mismatch", http.StatusBadRequest)
					return
				}
			}

			// Буферизация тела ответа
			rw := &responseWriterWithHash{
				ResponseWriter: w,
				buf:            &bytes.Buffer{},
			}

			next.ServeHTTP(rw, r)

			// Установка хеша тела ответа
			respHash := hashFunc(rw.buf.Bytes(), key)
			w.Header().Set(header, respHash)
			w.Write(rw.buf.Bytes())
		})
	}
}

// responseWriterWithHash — обёртка над http.ResponseWriter, которая буферизует тело ответа для последующего хеширования.
type responseWriterWithHash struct {
	http.ResponseWriter
	buf *bytes.Buffer
}

// Write записывает данные в буфер, не отправляя их сразу клиенту.
func (w *responseWriterWithHash) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}
