package facades

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/go-yandex-practicum/internal/hash"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestNewMetricFacade_AddsHTTPPrefix(t *testing.T) {
	t.Run("adds http:// when no scheme is present", func(t *testing.T) {
		rawAddress := "example.com"
		client := resty.New()
		facade := NewMetricFacade(client, rawAddress, "")
		assert.Equal(t, "http://example.com", facade.client.BaseURL)
	})

	t.Run("preserves http:// prefix", func(t *testing.T) {
		rawAddress := "http://example.com"
		client := resty.New()
		facade := NewMetricFacade(client, rawAddress, "")
		assert.Equal(t, "http://example.com", facade.client.BaseURL)
	})

	t.Run("preserves https:// prefix", func(t *testing.T) {
		rawAddress := "https://example.com"
		client := resty.New()
		facade := NewMetricFacade(client, rawAddress, "")
		assert.Equal(t, "https://example.com", facade.client.BaseURL)
	})
}

func TestMetricFacade_Updates(t *testing.T) {
	key := "secretkey"
	tests := []struct {
		name            string
		serverHandler   http.HandlerFunc
		request         []types.Metrics
		serverURL       string
		expectError     bool
		expectedErrPart string
		withKey         bool
	}{
		{
			name: "Success gauge metric with key",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/updates/", r.URL.Path)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				receivedHash := r.Header.Get(hash.Header)
				assert.NotEmpty(t, receivedHash)

				bodyBytes := readRequestBody(t, r)
				expectedHash := hash.HashWithKey(bodyBytes, key)
				assert.Equal(t, expectedHash, receivedHash)

				w.WriteHeader(http.StatusOK)
			},
			request: []types.Metrics{
				{
					MetricID: types.MetricID{
						ID:   "CPU",
						Type: types.GaugeMetricType,
					},
					Value: float64Ptr(99.5),
				},
			},
			expectError: false,
			withKey:     true,
		},
		{
			name: "Success counter metric without key",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/updates/", r.URL.Path)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Empty(t, r.Header.Get(hash.Header))
				w.WriteHeader(http.StatusOK)
			},
			request: []types.Metrics{
				{
					MetricID: types.MetricID{
						ID:   "Requests",
						Type: types.CounterMetricType,
					},
					Delta: int64Ptr(42),
				},
			},
			expectError: false,
			withKey:     false,
		},
		{
			name: "Server returns error",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "bad request", http.StatusBadRequest)
			},
			request: []types.Metrics{
				{
					MetricID: types.MetricID{
						ID:   "Errors",
						Type: types.CounterMetricType,
					},
					Delta: int64Ptr(5),
				},
			},
			expectError:     true,
			expectedErrPart: "error response from server",
			withKey:         false,
		},
		{
			name:          "Bad URL",
			serverHandler: nil,
			request: []types.Metrics{
				{
					MetricID: types.MetricID{
						ID:   "Timeouts",
						Type: types.GaugeMetricType,
					},
					Value: float64Ptr(0.1),
				},
			},
			serverURL:       "http://invalid-host.local",
			expectError:     true,
			expectedErrPart: "failed to send metrics",
			withKey:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverURL := tt.serverURL
			var server *httptest.Server
			if tt.serverHandler != nil {
				server = httptest.NewServer(tt.serverHandler)
				defer server.Close()
				serverURL = server.URL
			}

			client := resty.New()
			var facade *MetricFacade
			if tt.withKey {
				facade = NewMetricFacade(client, serverURL, key)
			} else {
				facade = NewMetricFacade(client, serverURL, "")
			}
			err := facade.Updates(context.Background(), tt.request)
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

func float64Ptr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}

func readRequestBody(t *testing.T, r *http.Request) []byte {
	t.Helper()
	body, err := io.ReadAll(r.Body)
	assert.NoError(t, err)
	r.Body.Close()
	return body
}
