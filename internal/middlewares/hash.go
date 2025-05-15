package middlewares

import (
	"bytes"
	"io"
	"net/http"

	"github.com/sbilibin2017/go-yandex-practicum/internal/hash"
)

func HashMiddleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if key == "" {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read request body", http.StatusInternalServerError)
				return
			}
			r.Body.Close()
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			receivedHash := r.Header.Get(hash.Header)
			if receivedHash != "" {
				computedHash := hash.HashWithKey(bodyBytes, key)
				if !hash.CompareHash(receivedHash, computedHash) {
					http.Error(w, "hash mismatch", http.StatusBadRequest)
					return
				}
			}

			rw := &responseWriterWithHash{
				ResponseWriter: w,
				buf:            &bytes.Buffer{},
			}

			next.ServeHTTP(rw, r)

			if key != "" {
				respHash := hash.HashWithKey(rw.buf.Bytes(), key)
				w.Header().Set(hash.Header, respHash)
			}

			w.Write(rw.buf.Bytes())
		})
	}
}

type responseWriterWithHash struct {
	http.ResponseWriter
	buf *bytes.Buffer
}

func (w *responseWriterWithHash) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}
