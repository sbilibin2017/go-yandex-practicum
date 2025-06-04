package middlewares

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func generateTestKeys(t *testing.T) (privateKeyPEM []byte, privateKey *rsa.PrivateKey) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	privateKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})
	return privateKeyPEM, priv
}

func TestCryptoMiddleware_DecryptSuccess(t *testing.T) {
	privPEM, privKey := generateTestKeys(t)

	// Создаем middleware с приватным ключом
	tmpFile := t.TempDir() + "/priv.pem"
	err := os.WriteFile(tmpFile, privPEM, 0600)
	assert.NoError(t, err)

	middleware := CryptoMiddleware(tmpFile)

	// Исходный текст
	plainText := []byte("secret data")

	// Шифруем с публичным ключом
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, &privKey.PublicKey, plainText)
	assert.NoError(t, err)

	// Base64 кодируем
	encoded := base64.StdEncoding.EncodeToString(cipherText)

	// Создаем запрос с base64 телом
	req := httptest.NewRequest("POST", "/", strings.NewReader(encoded))
	rr := httptest.NewRecorder()

	called := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, plainText, body)
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)
	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestCryptoMiddleware_InvalidBase64(t *testing.T) {
	privPEM, _ := generateTestKeys(t)

	tmpFile := t.TempDir() + "/priv.pem"
	err := os.WriteFile(tmpFile, privPEM, 0600)
	assert.NoError(t, err)

	middleware := CryptoMiddleware(tmpFile)

	req := httptest.NewRequest("POST", "/", strings.NewReader("%%%invalid_base64%%%"))
	rr := httptest.NewRecorder()

	called := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	handler.ServeHTTP(rr, req)
	assert.False(t, called, "handler should not be called on invalid base64")
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCryptoMiddleware_NoKey_PassesBodyAsIs(t *testing.T) {
	middleware := CryptoMiddleware("")

	plainText := "not encrypted body"

	req := httptest.NewRequest("POST", "/", strings.NewReader(plainText))
	rr := httptest.NewRecorder()

	called := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, plainText, string(body))
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)
	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestCryptoMiddleware_EmptyBody(t *testing.T) {
	middleware := CryptoMiddleware("")

	req := httptest.NewRequest("POST", "/", nil)
	rr := httptest.NewRecorder()

	called := false
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)
	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rr.Code)
}
