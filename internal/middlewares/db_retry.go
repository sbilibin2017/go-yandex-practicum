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

type responseBufferWriter struct {
	http.ResponseWriter
	buf         *bytes.Buffer
	statusCode  int
	wroteHeader bool
}

func (w *responseBufferWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.wroteHeader = true
}

func (w *responseBufferWriter) Write(data []byte) (int, error) {
	return w.buf.Write(data)
}
