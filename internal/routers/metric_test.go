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
		middleware int
	}{}

	loggingMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called.middleware++
			next.ServeHTTP(w, r)
		})
	}

	gzipMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called.middleware++
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
		gzipMiddleware,
	)

	tests := []struct {
		name       string
		method     string
		path       string
		expectFunc func()
	}{
		{
			name:   "POST /update/{type}/{name}/{value}",
			method: http.MethodPost,
			path:   "/update/counter/myCounter/42",
			expectFunc: func() {
				assert.True(t, called.updatePath)
			},
		},
		{
			name:   "POST /update/",
			method: http.MethodPost,
			path:   "/update/",
			expectFunc: func() {
				assert.True(t, called.updateBody)
			},
		},
		{
			name:   "GET /value/{type}/{name}",
			method: http.MethodGet,
			path:   "/value/gauge/myGauge",
			expectFunc: func() {
				assert.True(t, called.getPath)
			},
		},
		{
			name:   "POST /value/",
			method: http.MethodPost,
			path:   "/value/",
			expectFunc: func() {
				assert.True(t, called.getBody)
			},
		},
		{
			name:   "GET /",
			method: http.MethodGet,
			path:   "/",
			expectFunc: func() {
				assert.True(t, called.list)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called = struct {
				updatePath bool
				updateBody bool
				getPath    bool
				getBody    bool
				list       bool
				middleware int
			}{}

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			tt.expectFunc()
			assert.Equal(t, 2, called.middleware)
		})
	}
}
