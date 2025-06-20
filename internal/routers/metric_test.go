package routers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sbilibin2017/go-yandex-practicum/internal/routers"
	"github.com/stretchr/testify/assert"
)

func dummyHandler(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(name))
	}
}

func TestNewMetricRouter(t *testing.T) {
	router := routers.NewMetricRouter(
		dummyHandler("updatePath"),
		dummyHandler("updateBody"),
		dummyHandler("updatesBody"),
		dummyHandler("getPath"),
		dummyHandler("getBody"),
		dummyHandler("listAll"),
	)

	tests := []struct {
		name       string
		method     string
		target     string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "Update with path parameters",
			method:     http.MethodPost,
			target:     "/update/gauge/test/123",
			wantStatus: http.StatusOK,
			wantBody:   "updatePath",
		},
		{
			name:       "Update with type and name only",
			method:     http.MethodPost,
			target:     "/update/gauge/test",
			wantStatus: http.StatusOK,
			wantBody:   "updatePath",
		},
		{
			name:       "Update with body",
			method:     http.MethodPost,
			target:     "/update/",
			wantStatus: http.StatusOK,
			wantBody:   "updateBody",
		},
		{
			name:       "Bulk update",
			method:     http.MethodPost,
			target:     "/updates/",
			wantStatus: http.StatusOK,
			wantBody:   "updatesBody",
		},
		{
			name:       "Get value by path",
			method:     http.MethodGet,
			target:     "/value/gauge/test",
			wantStatus: http.StatusOK,
			wantBody:   "getPath",
		},
		{
			name:       "Get value by body",
			method:     http.MethodPost,
			target:     "/value/",
			wantStatus: http.StatusOK,
			wantBody:   "getBody",
		},
		{
			name:       "List all metrics",
			method:     http.MethodGet,
			target:     "/",
			wantStatus: http.StatusOK,
			wantBody:   "listAll",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.target, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)

			body := w.Body.String()
			assert.Equal(t, tt.wantBody, body)
		})
	}
}
