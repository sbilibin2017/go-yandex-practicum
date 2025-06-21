package middlewares

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
)

// CryptoOption is a functional option for configuring the CryptoMiddleware.
type CryptoOption func(*cryptoMiddleware) error

// cryptoMiddleware holds the runtime state for the middleware.
type cryptoMiddleware struct {
	privateKey *rsa.PrivateKey
}

// WithKeyPath returns a CryptoOption that sets the RSA private key path and loads the key.
func WithKeyPath(path string) CryptoOption {
	return func(mw *cryptoMiddleware) error {
		if path == "" {
			return nil
		}
		key, err := loadPrivateKey(path)
		if err != nil {
			return err
		}
		mw.privateKey = key
		return nil
	}
}

// CryptoMiddleware returns an HTTP middleware that decrypts the request body using the configured private key.
// If no private key is set, it passes requests unchanged.
func CryptoMiddleware(opts ...CryptoOption) (func(http.Handler) http.Handler, error) {
	mw := &cryptoMiddleware{}

	// Apply options to middleware instance
	for _, opt := range opts {
		if err := opt(mw); err != nil {
			return nil, err
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If no private key, skip decryption
			if mw.privateKey == nil {
				next.ServeHTTP(w, r)
				return
			}

			encBody, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body.Close()

			if len(encBody) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			cipherText, err := base64.StdEncoding.DecodeString(string(encBody))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			plainText, err := rsa.DecryptPKCS1v15(rand.Reader, mw.privateKey, cipherText)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(strings.NewReader(string(plainText)))
			r.ContentLength = int64(len(plainText))

			next.ServeHTTP(w, r)
		})
	}, nil
}

// loadPrivateKey loads an RSA private key from PEM file.
func loadPrivateKey(keyPath string) (*rsa.PrivateKey, error) {
	if keyPath == "" {
		return nil, nil
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}
