package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

func HashMiddleware(key string, header string) func(http.Handler) http.Handler {
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
				h := hmac.New(sha256.New, []byte(key))
				h.Write(bodyBytes)
				computedHash := hex.EncodeToString(h.Sum(nil))

				if !hmac.Equal([]byte(receivedHash), []byte(computedHash)) {
					http.Error(w, "hash mismatch", http.StatusBadRequest)
					return
				}
			}

			rw := &responseWriterWithHash{
				ResponseWriter: w,
				buf:            &bytes.Buffer{},
			}

			next.ServeHTTP(rw, r)

			hResp := hmac.New(sha256.New, []byte(key))
			hResp.Write(rw.buf.Bytes())
			respHash := hex.EncodeToString(hResp.Sum(nil))

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
