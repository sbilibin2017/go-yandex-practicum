package middlewares

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"go.uber.org/zap"
)

// DBRetryMiddleware возвращает HTTP middleware, который оборачивает обработчик в логику повторных попыток (retry).
//
// Использует переданную функцию withRetry для управления повторами вызова обработчика.
//
// Аргументы:
//   - withRetry: функция выполнения с retry (например, retrier.WithRetry).
//   - attempts: срез интервалов между попытками.
//   - isRetriableErrorFuncs: опциональные функции, определяющие, является ли ошибка повторяемой.
//
// Поведение:
//   - Если внутри обработчика произошла ошибка или паника, будет предпринята повторная попытка.
//   - Ответ, сгенерированный в последней попытке, будет записан в http.ResponseWriter.
//   - Если все попытки неудачны, в лог будет записана ошибка и возвращён последний ответ.
func DBRetryMiddleware(
	withRetry func(
		ctx context.Context,
		attempts []time.Duration,
		fn func(ctx context.Context) error,
		isRetriableErrorFuncs ...func(err error) bool,
	) error,
	attempts []time.Duration,
	isRetriableErrorFuncs ...func(err error) bool,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			var lastBuf bytes.Buffer
			var lastStatus int

			err := withRetry(ctx, attempts, func(ctx context.Context) error {
				buf := &bytes.Buffer{}
				rw := &responseBufferWriter{ResponseWriter: w, buf: buf}

				errCh := make(chan error, 1)

				go func() {
					defer func() {
						if rec := recover(); rec != nil {
							errCh <- errors.New("handler panic occurred")
						}
					}()

					next.ServeHTTP(rw, r.WithContext(ctx))
					errCh <- nil
				}()

				err := <-errCh

				lastBuf = *bytes.NewBuffer(buf.Bytes())
				if rw.statusCode != 0 {
					lastStatus = rw.statusCode
				} else {
					lastStatus = http.StatusOK
				}

				return err
			}, isRetriableErrorFuncs...)

			if err != nil {
				logger.Log.Error("Handler failed after retries", zap.Error(err))
				w.WriteHeader(lastStatus)
				_, _ = w.Write(lastBuf.Bytes())
				return
			}

			w.WriteHeader(lastStatus)
			_, _ = w.Write(lastBuf.Bytes())
		})
	}
}

// responseBufferWriter — вспомогательная структура, реализующая http.ResponseWriter,
// но буферизующая тело ответа в памяти, чтобы его можно было переиспользовать между попытками.
//
// Используется внутри DBRetryMiddleware.
type responseBufferWriter struct {
	http.ResponseWriter
	buf         *bytes.Buffer // буфер для тела ответа
	statusCode  int           // код статуса, если был установлен
	wroteHeader bool          // признак, что заголовок был записан
}

// WriteHeader сохраняет HTTP статус, но не отправляет его немедленно.
func (w *responseBufferWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.wroteHeader = true
}

// Write записывает тело ответа в буфер, не отправляя его клиенту напрямую.
func (w *responseBufferWriter) Write(data []byte) (int, error) {
	return w.buf.Write(data)
}
