package middlewares

import (
	"net/http"
	"time"

	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"go.uber.org/zap"
)

// LoggingMiddleware — HTTP middleware, логирующий информацию о каждом запросе и ответе.
//
// Логирует:
//   - URI и метод запроса
//   - Время выполнения запроса (duration)
//   - HTTP-статус ответа
//   - Размер тела ответа (в байтах)
//
// Пример использования:
//
//	http.Handle("/", LoggingMiddleware(myHandler))
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		logger.Log.Info("Request info",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Duration("duration", duration),
		)

		logger.Log.Info("Response info",
			zap.Int("status", rw.statusCode),
			zap.Int("size", rw.size),
		)
	})
}

// responseWriter — обёртка над http.ResponseWriter для перехвата статуса и размера ответа.
//
// Используется в LoggingMiddleware для логирования информации об ответе.
type responseWriter struct {
	http.ResponseWriter
	statusCode int // HTTP-статус ответа
	size       int // Объём тела ответа в байтах
}

// WriteHeader сохраняет статус ответа и передаёт его оригинальному ResponseWriter.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write записывает тело ответа и сохраняет его размер.
func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}
