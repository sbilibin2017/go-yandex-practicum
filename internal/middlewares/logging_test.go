package middlewares

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// testHandler is a simple handler that writes status and response body.
func testHandler(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func TestLoggingMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		handlerStatus  int
		handlerBody    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "200 OK with body",
			handlerStatus:  http.StatusOK,
			handlerBody:    "hello world",
			expectedStatus: http.StatusOK,
			expectedBody:   "hello world",
		},
		{
			name:           "404 Not Found with empty body",
			handlerStatus:  http.StatusNotFound,
			handlerBody:    "",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
		},
		{
			name:           "500 Internal Server Error with body",
			handlerStatus:  http.StatusInternalServerError,
			handlerBody:    "error occurred",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := LoggingMiddleware(testHandler(tt.handlerStatus, tt.handlerBody))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			h.ServeHTTP(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			require.Equal(t, tt.expectedStatus, resp.StatusCode)
			bodyBytes, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, tt.expectedBody, string(bodyBytes))
		})
	}
}
