package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

// HashMiddlewareConfig holds configuration for the HashMiddleware.
//
// Key is the secret key used for HMAC calculation.
// Header specifies the HTTP header used for sending/verifying hashes.
type HashMiddlewareConfig struct {
	Key    string
	Header string
}

// HashMiddlewareOption defines a functional option for configuring HashMiddlewareConfig.
type HashMiddlewareOption func(*HashMiddlewareConfig)

// NewHashMiddlewareConfig creates a new HashMiddlewareConfig applying the given options.
func NewHashMiddlewareConfig(opts ...HashMiddlewareOption) (*HashMiddlewareConfig, error) {
	cfg := &HashMiddlewareConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg, nil
}

// WithHashKey sets the secret key for the HashMiddlewareConfig.
func WithHashKey(key string) HashMiddlewareOption {
	return func(cfg *HashMiddlewareConfig) {
		cfg.Key = key
	}
}

// WithHashHeader sets the HTTP header name used to read and write hashes.
func WithHashHeader(header string) HashMiddlewareOption {
	return func(cfg *HashMiddlewareConfig) {
		cfg.Header = header
	}
}

// HashMiddleware returns an HTTP middleware that verifies incoming request body HMAC SHA256 hash
// from the configured header and adds a response body HMAC SHA256 hash to the same header.
//
// If the Key is empty, the middleware skips all processing and returns the next handler unchanged.
func HashMiddleware(opts ...HashMiddlewareOption) (func(http.Handler) http.Handler, error) {
	cfg, err := NewHashMiddlewareConfig(opts...)
	if err != nil {
		return nil, err
	}

	return func(next http.Handler) http.Handler {
		if cfg.Key == "" {
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

			receivedHash := r.Header.Get(cfg.Header)
			if receivedHash != "" {
				mac := hmac.New(sha256.New, []byte(cfg.Key))
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

			mac := hmac.New(sha256.New, []byte(cfg.Key))
			mac.Write(rw.buf.Bytes())
			respMAC := mac.Sum(nil)
			respHash := hex.EncodeToString(respMAC)

			w.Header().Set(cfg.Header, respHash)
			w.Write(rw.buf.Bytes())
		})
	}, nil
}

// responseWriterWithHash captures response body data for hash calculation.
type responseWriterWithHash struct {
	http.ResponseWriter
	buf *bytes.Buffer
}

func (w *responseWriterWithHash) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}
