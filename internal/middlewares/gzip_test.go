package middlewares

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGzipMiddleware_InvalidGzipRequest(t *testing.T) {
	invalidGzipBody := bytes.NewBufferString("this is not gzipped data")
	req := httptest.NewRequest("POST", "/bad-gzip", invalidGzipBody)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler := GzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called when gzip decompression fails")
	}))
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Failed to read gzip data")
}

func TestGzipMiddleware_DecompressRequest(t *testing.T) {
	data := []byte(`{"message": "Hello, World!"}`)
	var compressedData bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedData)
	_, err := gzipWriter.Write(data)
	assert.NoError(t, err)
	gzipWriter.Close()
	req := httptest.NewRequest("POST", "/json", &compressedData)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler := GzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.JSONEq(t, `{"message": "Hello, World!"}`, string(body))
		w.WriteHeader(http.StatusOK)
	}))
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGzipMiddleware_CompressResponse(t *testing.T) {
	req := httptest.NewRequest("GET", "/json", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler := GzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Hello, World!"}`))
	}))
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "gzip", rr.Header().Get("Content-Encoding"))
	gzipReader, err := gzip.NewReader(rr.Body)
	assert.NoError(t, err)
	defer gzipReader.Close()
	uncompressedData, err := io.ReadAll(gzipReader)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"message": "Hello, World!"}`, string(uncompressedData))
}
