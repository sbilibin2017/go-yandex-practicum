package middlewares

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipMiddleware_CompressesResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	rr := httptest.NewRecorder()

	middleware := GzipMiddleware(handler)
	middleware.ServeHTTP(rr, req)

	assert.Equal(t, "gzip", rr.Header().Get("Content-Encoding"))

	gr, err := gzip.NewReader(rr.Body)
	require.NoError(t, err)
	defer gr.Close()

	body, err := io.ReadAll(gr)
	require.NoError(t, err)
	assert.Equal(t, "Hello, world!", string(body))
}

func TestGzipMiddleware_DecompressesRequest(t *testing.T) {
	var receivedBody string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		receivedBody = string(body)
		w.Write([]byte("ok"))
	})

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write([]byte("compressed data"))
	require.NoError(t, err)
	zw.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")

	rr := httptest.NewRecorder()
	middleware := GzipMiddleware(handler)
	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "compressed data", receivedBody)
}

func TestGzipMiddleware_InvalidGzipRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Handler should not be called on invalid gzip")
	})

	invalidBody := strings.NewReader("not gzip")
	req := httptest.NewRequest(http.MethodPost, "/", invalidBody)
	req.Header.Set("Content-Encoding", "gzip")

	rr := httptest.NewRecorder()
	middleware := GzipMiddleware(handler)
	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
