package middlewares

import (
	"bytes"
	"io"
	"net/http"
)

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
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read request body", http.StatusInternalServerError)
				return
			}
			r.Body.Close()
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			receivedHash := r.Header.Get(header)
			if receivedHash != "" {
				computedHash := hashFunc(bodyBytes, key)

				if !compareFunc(receivedHash, computedHash) {
					http.Error(w, "hash mismatch", http.StatusBadRequest)
					return
				}
			}

			rw := &responseWriterWithHash{
				ResponseWriter: w,
				buf:            &bytes.Buffer{},
			}

			next.ServeHTTP(rw, r)

			respHash := hashFunc(rw.buf.Bytes(), key)
			w.Header().Set(header, respHash)
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
