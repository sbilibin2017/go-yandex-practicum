package middlewares

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sbilibin2017/go-yandex-practicum/internal/hash"
	"github.com/stretchr/testify/assert"
)

func TestHashMiddleware_NoKey(t *testing.T) {
	handler := HashMiddleware("")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	handler(nextHandler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func TestHashMiddleware_CorrectHash(t *testing.T) {
	key := "secret"
	body := []byte(`{"metric":"value"}`)
	hashVal := hash.HashWithKey(body, key)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(hash.Header, hashVal)
	w := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`response data`))
	})

	handler := HashMiddleware(key)
	handler(nextHandler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	respBody := w.Body.Bytes()
	assert.Equal(t, "response data", string(respBody))

	respHash := w.Header().Get(hash.Header)
	expectedRespHash := hash.HashWithKey(respBody, key)
	assert.Equal(t, expectedRespHash, respHash)
}

func TestHashMiddleware_IncorrectHash(t *testing.T) {
	key := "secret"
	body := []byte(`{"metric":"value"}`)
	incorrectHash := "bad_hash"

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(hash.Header, incorrectHash)
	w := httptest.NewRecorder()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("should not be called"))
	})

	handler := HashMiddleware(key)
	handler(nextHandler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "hash mismatch")
}
