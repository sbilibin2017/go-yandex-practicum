package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type brokenReader struct{}

func (b *brokenReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestHashMiddleware_BodyReadError(t *testing.T) {
	key := "secretkey"
	header := "X-Hash"

	brokenBody := &brokenReader{}

	handler := HashMiddleware(key, header)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called if body read fails")
	}))

	req := httptest.NewRequest(http.MethodPost, "/", brokenBody)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Body will be empty now, since middleware does not write error message
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Len(t, respBody, 0, "response body should be empty on body read error")
}

func TestHashMiddleware_EmptyKey_ReturnsNext(t *testing.T) {
	key := ""
	header := "X-Hash"

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := HashMiddleware(key, header)(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.True(t, called, "Next handler should be called")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHashMiddleware_ValidHash(t *testing.T) {
	key := "secretkey"
	header := "X-Hash"

	handler := HashMiddleware(key, header)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		w.Write(bodyBytes)
	}))

	body := []byte(`{"foo":"bar"}`)

	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(body)
	expectedMAC := mac.Sum(nil)
	expectedHash := hex.EncodeToString(expectedMAC)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(header, expectedHash)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, string(body), string(respBody))
	assert.Equal(t, expectedHash, resp.Header.Get(header))
}

func TestHashMiddleware_MissingHash(t *testing.T) {
	key := "secretkey"
	header := "X-Hash"

	handler := HashMiddleware(key, header)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		w.Write(bodyBytes)
	}))

	body := []byte(`{"foo":"bar"}`)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, string(body), string(respBody))

	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(body)
	expectedHash := hex.EncodeToString(mac.Sum(nil))
	assert.Equal(t, expectedHash, resp.Header.Get(header))
}

func TestHashMiddleware_InvalidHash(t *testing.T) {
	key := "secretkey"
	header := "X-Hash"

	handler := HashMiddleware(key, header)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called if hash mismatch")
	}))

	body := []byte(`{"foo":"bar"}`)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(header, "invalidhashvalue")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Len(t, respBody, 0, "response body should be empty on hash mismatch")
}
