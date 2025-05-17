package facades

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestNewMetricFacade_AddsHTTPPrefix(t *testing.T) {
	client := resty.New()

	// Адрес без http/https
	addr := "example.com"
	facade := NewMetricFacade(client, addr, "key", "HashSHA256")

	// Проверяем, что у клиента установлен базовый URL с префиксом http://
	assert.True(t, facade.client.BaseURL != "", "BaseURL should be set")

	baseURL := facade.client.BaseURL
	assert.True(t, strings.HasPrefix(baseURL, "http://"), "BaseURL should start with http://")

	// Если адрес уже с префиксом https://
	addrHTTPS := "https://secure.com"
	facade2 := NewMetricFacade(client, addrHTTPS, "key", "HashSHA256")
	assert.Equal(t, addrHTTPS, facade2.client.BaseURL)
}

func TestMetricFacade_Updates(t *testing.T) {
	key := "secretkey"
	headerName := "HashSHA256"

	tests := []struct {
		name            string
		serverHandler   http.HandlerFunc
		request         []types.Metrics
		serverURL       string
		expectError     bool
		expectedErrPart string
		key             string
		header          string
	}{
		{
			name: "Success gauge metric with key",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				assert.Equal(t, "/updates/", r.URL.Path)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				receivedHash := r.Header.Get(headerName)
				assert.NotEmpty(t, receivedHash)

				bodyBytes := readRequestBody(t, r)

				// Вычисляем хэш напрямую (вместо вызова hash.HashWithKey)
				h := hmac.New(sha256.New, []byte(key))
				h.Write(bodyBytes)
				expectedHash := hex.EncodeToString(h.Sum(nil))

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
			key:         key,
			header:      headerName,
		},
		{
			name: "Success counter metric without key",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				assert.Equal(t, "/updates/", r.URL.Path)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Empty(t, r.Header.Get(headerName))
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
			key:         "",
			header:      headerName,
		},
		{
			name: "Server returns error",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
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
			key:             "",
			header:          headerName,
		},
		{
			name:      "Bad URL",
			serverURL: "http://invalid-host.local",
			request: []types.Metrics{
				{
					MetricID: types.MetricID{
						ID:   "Timeouts",
						Type: types.GaugeMetricType,
					},
					Value: float64Ptr(0.1),
				},
			},
			expectError:     true,
			expectedErrPart: "failed to send metrics",
			key:             "",
			header:          headerName,
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
			facade := NewMetricFacade(client, serverURL, tt.key, tt.header)

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
	return body
}

func TestMetricFacade_Updates_EmptyMetrics(t *testing.T) {
	client := resty.New()
	facade := NewMetricFacade(client, "http://localhost", "key", "HashSHA256")
	err := facade.Updates(context.Background(), []types.Metrics{})
	assert.NoError(t, err)
}
