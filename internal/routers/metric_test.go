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
		update bool
		get    bool
		list   bool
		mw     bool
	}{}

	// mock middleware: просто отмечает, что был вызван
	loggingMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called.mw = true
			next.ServeHTTP(w, r)
		})
	}

	// mock handlers
	updateHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called.update = true
		w.WriteHeader(http.StatusOK)
	})
	getHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called.get = true
		w.WriteHeader(http.StatusOK)
	})
	listHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called.list = true
		w.WriteHeader(http.StatusOK)
	})

	router := NewMetricRouter(updateHandler, getHandler, listHandler, loggingMiddleware)

	t.Run("POST /update/{type}/{name}/{value}", func(t *testing.T) {
		called = struct{ update, get, list, mw bool }{}
		req := httptest.NewRequest(http.MethodPost, "/update/gauge/testmetric/123.4", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, called.update)
		assert.True(t, called.mw)
	})

	t.Run("GET /value/{type}/{name}", func(t *testing.T) {
		called = struct{ update, get, list, mw bool }{}
		req := httptest.NewRequest(http.MethodGet, "/value/counter/testmetric", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, called.get)
		assert.True(t, called.mw)
	})

	t.Run("GET /", func(t *testing.T) {
		called = struct{ update, get, list, mw bool }{}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, called.list)
		assert.True(t, called.mw)
	})
}
