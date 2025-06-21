package middlewares

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper to create temp private key file for tests
func createTempPrivateKeyFile(t *testing.T) string {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privBytes := x509.MarshalPKCS1PrivateKey(privKey)
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	}

	tmpFile, err := os.CreateTemp("", "testkey_*.pem")
	require.NoError(t, err)

	// Correctly encode PEM block to the file
	err = pem.Encode(tmpFile, pemBlock)
	require.NoError(t, err)

	err = tmpFile.Close()
	require.NoError(t, err)

	return tmpFile.Name()
}

func TestCryptoMiddleware(t *testing.T) {
	keyPath := createTempPrivateKeyFile(t)
	defer os.Remove(keyPath)

	privKey, err := loadPrivateKey(keyPath)
	require.NoError(t, err)
	pubKey := &privKey.PublicKey

	plaintext := "test message"

	// encrypt plaintext with public key and base64 encode
	cipherBytes, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, []byte(plaintext))
	require.NoError(t, err)

	encBody := base64.StdEncoding.EncodeToString(cipherBytes)

	type args struct {
		keyPath  string
		body     []byte
		expected string
		status   int
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "decrypts valid encrypted body",
			args: args{
				keyPath:  keyPath,
				body:     []byte(encBody),
				expected: plaintext,
				status:   http.StatusOK,
			},
		},
		{
			name: "passes through empty body",
			args: args{
				keyPath:  keyPath,
				body:     []byte(""),
				expected: "",
				status:   http.StatusOK,
			},
		},
		{
			name: "bad base64 returns 400",
			args: args{
				keyPath:  keyPath,
				body:     []byte("invalid-base64$$"),
				expected: "",
				status:   http.StatusBadRequest,
			},
		},
		{
			name: "no key path disables decryption and passes body as is",
			args: args{
				keyPath:  "",
				body:     []byte(plaintext),
				expected: plaintext,
				status:   http.StatusOK,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware, err := CryptoMiddleware(WithKeyPath(tt.args.keyPath))
			require.NoError(t, err)

			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				bodyBytes, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				w.Write(bodyBytes)
			}))

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(tt.args.body))
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			resp := rec.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.args.status, resp.StatusCode)

			if tt.args.status == http.StatusOK {
				respBody, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.args.expected, string(respBody))
			}
		})
	}
}
