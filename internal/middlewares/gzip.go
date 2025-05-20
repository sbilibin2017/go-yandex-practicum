package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"go.uber.org/zap"
)

// GzipMiddleware — HTTP middleware, обеспечивающее поддержку gzip-сжатия для входящих и исходящих HTTP-запросов.
//
// Поведение:
//   - Если Content-Encoding запроса — "gzip", тело запроса распаковывается.
//   - Если клиент поддерживает gzip (в заголовке Accept-Encoding), ответ сжимается и отсылается с заголовком Content-Encoding: gzip.
//
// Используется для уменьшения объема передаваемых данных между клиентом и сервером.
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Распаковка тела запроса, если оно было сжато gzip
		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				logger.Log.Error("Failed to read gzip data from request", zap.Error(err))
				http.Error(w, "Failed to read gzip data", http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = gz
		}

		// Сжатие ответа, если клиент это поддерживает
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			defer gz.Close()
			w = &gzipResponseWriter{Writer: gz, ResponseWriter: w}
		}

		next.ServeHTTP(w, r)
	})
}

// gzipResponseWriter — обёртка над http.ResponseWriter, осуществляющая gzip-сжатие тела ответа.
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

// Write реализует интерфейс http.ResponseWriter, сжимая данные перед их отправкой клиенту.
func (grw *gzipResponseWriter) Write(p []byte) (n int, err error) {
	return grw.Writer.Write(p)
}
