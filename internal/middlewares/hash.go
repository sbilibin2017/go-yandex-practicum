package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

// HashOption is a functional option for configuring the hash middleware.
type HashOption func(*hashMiddleware)

// hashMiddleware holds middleware runtime configuration.
type hashMiddleware struct {
	key    string
	header string
}

// WithHashKey sets the secret key for HMAC.
func WithHashKey(key string) HashOption {
	return func(mw *hashMiddleware) {
		mw.key = key
	}
}

// WithHashHeader sets the HTTP header used to send/verify hashes.
func WithHashHeader(header string) HashOption {
	return func(mw *hashMiddleware) {
		mw.header = header
	}
}

// HashMiddleware returns a middleware handler that verifies request body HMAC SHA256 and
// adds response body HMAC SHA256 in the configured header.
// If the key is empty, the middleware skips all processing.
func HashMiddleware(opts ...HashOption) (func(http.Handler) http.Handler, error) {
	mw := &hashMiddleware{}
	for _, opt := range opts {
		opt(mw)
	}
	return func(next http.Handler) http.Handler {
		if mw.key == "" {
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

			receivedHash := r.Header.Get(mw.header)
			if receivedHash != "" {
				mac := hmac.New(sha256.New, []byte(mw.key))
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

			mac := hmac.New(sha256.New, []byte(mw.key))
			mac.Write(rw.buf.Bytes())
			respMAC := mac.Sum(nil)
			respHash := hex.EncodeToString(respMAC)

			w.Header().Set(mw.header, respHash)
			w.Write(rw.buf.Bytes())
		})
	}, nil
}

// responseWriterWithHash captures the response body for hash calculation.
type responseWriterWithHash struct {
	http.ResponseWriter
	buf *bytes.Buffer
}

func (w *responseWriterWithHash) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}
