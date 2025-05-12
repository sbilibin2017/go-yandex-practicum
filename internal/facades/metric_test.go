package facades_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"

	"github.com/sbilibin2017/go-yandex-practicum/internal/facades"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestMetricFacade_Update(t *testing.T) {
	tests := []struct {
		name            string
		serverHandler   http.HandlerFunc
		request         types.MetricUpdatePathRequest
		serverURL       string
		expectError     bool
		expectedErrPart string
	}{
		{
			name: "Success response",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/update/gauge/CPU/99.5", r.URL.Path)
				assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))
				w.WriteHeader(http.StatusOK)
			},
			request: types.MetricUpdatePathRequest{
				Type:  "gauge",
				Name:  "CPU",
				Value: "99.5",
			},
			expectError: false,
		},
		{
			name: "Server returns error response",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "bad request", http.StatusBadRequest)
			},
			request: types.MetricUpdatePathRequest{
				Type:  "counter",
				Name:  "Requests",
				Value: "10",
			},
			expectError:     true,
			expectedErrPart: "error response from server",
		},
		{
			name:            "Request error due to bad URL",
			serverHandler:   nil,
			request:         types.MetricUpdatePathRequest{Type: "counter", Name: "Errors", Value: "5"},
			serverURL:       "http://invalid-host",
			expectError:     true,
			expectedErrPart: "failed to send metric",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var serverURL string
			if tt.serverHandler != nil {
				server := httptest.NewServer(tt.serverHandler)
				defer server.Close()
				serverURL = server.URL
			} else {
				serverURL = tt.serverURL
			}

			client := resty.New()
			facade := facades.NewMetricFacade(*client, serverURL)

			err := facade.Update(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedErrPart != "" {
					assert.Contains(t, err.Error(), tt.expectedErrPart)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
