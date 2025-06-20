package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

func HashMiddleware(
	key string,
	header string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if key == "" || header == "" {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body.Close()
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			receivedHash := r.Header.Get(header)
			if receivedHash != "" {
				mac := hmac.New(sha256.New, []byte(key))
				mac.Write(bodyBytes)
				expectedMAC := mac.Sum(nil)
				expectedHash := hex.EncodeToString(expectedMAC)

				if !hmac.Equal([]byte(receivedHash), []byte(expectedHash)) {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			}

			rw := &responseWriterWithHash{
				ResponseWriter: w,
				buf:            &bytes.Buffer{},
			}

			next.ServeHTTP(rw, r)

			mac := hmac.New(sha256.New, []byte(key))
			mac.Write(rw.buf.Bytes())
			respMAC := mac.Sum(nil)
			respHash := hex.EncodeToString(respMAC)

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
