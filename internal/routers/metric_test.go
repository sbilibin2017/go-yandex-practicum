package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricRouter(t *testing.T) {
	called := struct {
		updatePath bool
		updateBody bool
		getPath    bool
		getBody    bool
		list       bool
		middleware bool
	}{}

	loggingMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called.middleware = true
			next.ServeHTTP(w, r)
		})
	}

	updatePathHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called.updatePath = true
		w.WriteHeader(http.StatusOK)
	})
	updateBodyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called.updateBody = true
		w.WriteHeader(http.StatusOK)
	})
	getPathHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called.getPath = true
		w.WriteHeader(http.StatusOK)
	})
	getBodyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called.getBody = true
		w.WriteHeader(http.StatusOK)
	})
	listHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called.list = true
		w.WriteHeader(http.StatusOK)
	})

	router := NewMetricRouter(
		updatePathHandler,
		updateBodyHandler,
		getPathHandler,
		getBodyHandler,
		listHandler,
		loggingMiddleware,
	)

	t.Run("POST /update/{type}/{name}/{value}", func(t *testing.T) {
		called = struct{ updatePath, updateBody, getPath, getBody, list, middleware bool }{}
		req := httptest.NewRequest(http.MethodPost, "/update/counter/myCounter/42", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, called.updatePath)
		assert.True(t, called.middleware)
	})

	t.Run("POST /update/", func(t *testing.T) {
		called = struct{ updatePath, updateBody, getPath, getBody, list, middleware bool }{}
		req := httptest.NewRequest(http.MethodPost, "/update/", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, called.updateBody)
		assert.True(t, called.middleware)
	})

	t.Run("GET /value/{type}/{name}", func(t *testing.T) {
		called = struct{ updatePath, updateBody, getPath, getBody, list, middleware bool }{}
		req := httptest.NewRequest(http.MethodGet, "/value/gauge/myGauge", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, called.getPath)
		assert.True(t, called.middleware)
	})

	t.Run("POST /value/", func(t *testing.T) {
		called = struct{ updatePath, updateBody, getPath, getBody, list, middleware bool }{}
		req := httptest.NewRequest(http.MethodPost, "/value/", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, called.getBody)
		assert.True(t, called.middleware)
	})

	t.Run("GET /", func(t *testing.T) {
		called = struct{ updatePath, updateBody, getPath, getBody, list, middleware bool }{}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, called.list)
		assert.True(t, called.middleware)
	})
}
