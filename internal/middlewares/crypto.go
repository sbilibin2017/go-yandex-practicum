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

// CryptoMiddlewareConfig holds configuration options for the crypto middleware.
// It includes the path to the RSA private key and a cached parsed PrivateKey.
type CryptoMiddlewareConfig struct {
	KeyPath    string
	PrivateKey *rsa.PrivateKey // internally cached after loading KeyPath
}

// CryptoOption defines a functional option for configuring CryptoMiddlewareConfig.
type CryptoOption func(*CryptoMiddlewareConfig) error

// NewCryptoMiddlewareConfig creates a new CryptoMiddlewareConfig by applying
// the given functional options and loading the private key if KeyPath is set.
// Returns an error if key loading fails.
func NewCryptoMiddlewareConfig(opts ...CryptoOption) (*CryptoMiddlewareConfig, error) {
	options := &CryptoMiddlewareConfig{}

	// Apply functional options
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return nil, err
		}
	}

	// Load private key after options applied
	key, err := loadPrivateKey(options.KeyPath)
	if err != nil {
		return nil, err
	}

	options.PrivateKey = key
	return options, nil
}

// CryptoMiddleware returns an HTTP middleware handler that decrypts the
// incoming request body if a private key is configured.
// It reads the encrypted request body (base64-encoded), decrypts it using RSA PKCS#1 v1.5,
// and replaces the request body with the decrypted plaintext.
// If no key path is set, it passes the request through without changes.
//
// Returns an error if the configuration cannot be created or key cannot be loaded.
func CryptoMiddleware(opts ...CryptoOption) (func(http.Handler) http.Handler, error) {
	config, err := NewCryptoMiddlewareConfig(opts...)
	if err != nil {
		return nil, err
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.KeyPath == "" {
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

			var plainText []byte
			if config.PrivateKey != nil {
				cipherText, err := base64.StdEncoding.DecodeString(string(encBody))
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				plainText, err = rsa.DecryptPKCS1v15(rand.Reader, config.PrivateKey, cipherText)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			} else {
				plainText = encBody
			}

			r.Body = io.NopCloser(strings.NewReader(string(plainText)))
			r.ContentLength = int64(len(plainText))

			next.ServeHTTP(w, r)
		})
	}, nil
}

// loadPrivateKey loads an RSA private key from the PEM file specified by keyPath.
// Returns nil if keyPath is empty.
// Returns an error if the file cannot be read or the key cannot be parsed.
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
		err := errors.New("failed to decode PEM block containing private key")
		return nil, err
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// WithKeyPath returns a CryptoOption to set the KeyPath in CryptoMiddlewareConfig.
func WithKeyPath(path string) CryptoOption {
	return func(cfg *CryptoMiddlewareConfig) error {
		cfg.KeyPath = path
		return nil
	}
}
