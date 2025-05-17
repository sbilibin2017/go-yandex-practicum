package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashMiddleware_NoKey(t *testing.T) {
	key := ""
	headerName := "X-Hash"
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	handler := HashMiddleware(key, headerName)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	handler(nextHandler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
	// При пустом ключе заголовок не должен устанавливаться
	assert.Empty(t, w.Header().Get(headerName))
}

func TestHashMiddleware_CorrectHash(t *testing.T) {
	key := "secret"
	headerName := "X-Hash"
	body := []byte(`{"metric":"value"}`)

	// Вычисляем корректный хэш с тем же кодом, что в middleware
	h := hmac.New(sha256.New, []byte(key))
	h.Write(body)
	correctHash := hex.EncodeToString(h.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(headerName, correctHash)

	w := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("response data"))
	})

	handler := HashMiddleware(key, headerName)
	handler(nextHandler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "response data", w.Body.String())

	// Проверяем, что заголовок с хэшем ответа установлен корректно
	respBody := w.Body.Bytes()
	hResp := hmac.New(sha256.New, []byte(key))
	hResp.Write(respBody)
	expectedRespHash := hex.EncodeToString(hResp.Sum(nil))

	assert.Equal(t, expectedRespHash, w.Header().Get(headerName))
}

func TestHashMiddleware_IncorrectHash(t *testing.T) {
	key := "secret"
	headerName := "X-Hash"
	body := []byte(`{"metric":"value"}`)
	incorrectHash := "bad_hash"

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(headerName, incorrectHash)

	w := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("nextHandler should not be called on hash mismatch")
	})

	handler := HashMiddleware(key, headerName)
	handler(nextHandler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "hash mismatch")
}

func TestHashMiddleware_ReadBodyError(t *testing.T) {
	key := "secret"
	headerName := "X-Hash"
	brokenReader := &errReader{}

	req := httptest.NewRequest(http.MethodPost, "/", brokenReader)
	w := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("nextHandler should not be called when body read fails")
	})

	handler := HashMiddleware(key, headerName)
	handler(nextHandler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to read request body")
}

type errReader struct{}

func (r *errReader) Read(p []byte) (int, error) {
	return 0, errors.New("read error")
}
