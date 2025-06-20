package http

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareServerAddress(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"example.com", "http://example.com"},
		{"http://example.com", "http://example.com"},
		{"https://secure.com", "https://secure.com"},
		{"", "http://"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := prepareServerAddress(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPrepareURL(t *testing.T) {
	tests := []struct {
		serverAddress, serverEndpoint, want string
	}{
		{"http://example.com", "/path/", "http://example.com/path/"},
		{"http://example.com/", "path", "http://example.com/path"},
		{"http://example.com/", "/path", "http://example.com/path"},
		{"http://example.com", "path", "http://example.com/path"},
	}
	for _, tt := range tests {
		t.Run(tt.serverAddress+" + "+tt.serverEndpoint, func(t *testing.T) {
			got := prepareURL(tt.serverAddress, tt.serverEndpoint)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPrepareRequest(t *testing.T) {
	val := 123.456
	delta := int64(789)
	metrics := []*types.Metrics{
		{ID: "1", MType: "gauge", Value: &val},
		{ID: "2", MType: "counter", Delta: &delta},
	}

	// Without encryption
	body, hashSum, err := prepareRequest(metrics, "key123", "")
	require.NoError(t, err)
	require.NotEmpty(t, body)
	require.NotEmpty(t, hashSum)

	// With encryption - prepare public key file
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	pubASN1, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubASN1})

	tmpFile, err := os.CreateTemp("", "pubkey*.pem")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write(pubPEM)
	require.NoError(t, err)
	tmpFile.Close()

	bodyEnc, hashSumEnc, err := prepareRequest(metrics, "key123", tmpFile.Name())
	require.NoError(t, err)
	require.NotEmpty(t, bodyEnc)
	require.NotEmpty(t, hashSumEnc)
	assert.NotEqual(t, body, bodyEnc)

	// With invalid public key path
	_, _, err = prepareRequest(metrics, "key123", "/non/existent/path.pem")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load public key")
}

func TestCalcBodyHashSum(t *testing.T) {
	body := []byte("hello world")
	key := "secret"
	sum := calcBodyHashSum(body, key)
	require.NotEmpty(t, sum)
}

func TestExtractIP(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    string
	}{
		{
			name:    "Full HTTP URL with IP and port",
			address: "http://127.0.0.1:8080/path",
			want:    "127.0.0.1",
		},
		{
			name:    "Full HTTPS URL with hostname and port",
			address: "https://example.com:443/path",
			want:    "example.com",
		},
		{
			name:    "URL without port",
			address: "http://example.com/path",
			want:    "example.com",
		},
		{
			name:    "Just IP with port",
			address: "192.168.1.1:9000",
			want:    "192.168.1.1",
		},
		{
			name:    "Just IP without port",
			address: "10.0.0.1",
			want:    "10.0.0.1",
		},
		{
			name:    "Localhost IP should return empty",
			address: "http://localhost:8080",
			want:    "",
		},
		{
			name:    "Empty address returns empty",
			address: "",
			want:    "",
		},
		{
			name:    "Malformed URL returns empty",
			address: "://bad_url",
			want:    "",
		},
		{
			name:    "Hostname localhost returns empty",
			address: "localhost",
			want:    "",
		},
		{
			name:    "IPv6 address with port",
			address: "http://[2001:db8::1]:8080",
			want:    "2001:db8::1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractIP(tt.address)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSendRequest(t *testing.T) {
	// Successful request test
	t.Run("Success", func(t *testing.T) {
		// Start test server
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "gzip", r.Header.Get("Accept-Encoding"))
			assert.NotEmpty(t, r.Header.Get("X-Real-IP")) // IP extracted from URL
			assert.Equal(t, "hashsumvalue", r.Header.Get("X-Test-Header"))

			// You can also check the body if needed here

			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		client := resty.New()

		err := sendRequest(
			context.Background(),
			client,
			ts.URL,
			[]byte(`{"foo":"bar"}`),
			"X-Test-Header",
			"hashsumvalue",
		)
		require.NoError(t, err)
	})

	// Server returns error code
	t.Run("ServerError", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "bad request", http.StatusBadRequest)
		}))
		defer ts.Close()

		client := resty.New()

		err := sendRequest(
			context.Background(),
			client,
			ts.URL,
			[]byte(`{}`),
			"X-Test-Header",
			"hashsumvalue",
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "server error 400")
	})

	// Invalid URL simulating client error
	t.Run("ClientError", func(t *testing.T) {
		client := resty.New()
		client.SetTimeout(1 * time.Second) // short timeout to avoid hanging

		err := sendRequest(
			context.Background(),
			client,
			"http://invalid.host", // unreachable host
			[]byte(`{}`),
			"X-Test-Header",
			"hashsumvalue",
		)
		require.Error(t, err)
	})
}

func sampleMetrics() []*types.Metrics {
	val := 123.45
	delta := int64(10)
	return []*types.Metrics{
		{
			ID:    "metric1",
			MType: "gauge",
			Value: &val,
			Delta: &delta,
		},
	}
}

func TestMetricsFacade_Updates(t *testing.T) {
	metrics := sampleMetrics()

	// Mock server to verify request received by Updates method
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Accept-Encoding"))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	facade, err := NewMetricsFacade(ts.URL, "X-Custom-Header", "secretkey", "")
	require.NoError(t, err)

	// Replace internal resty client with one having timeout to avoid hangs
	facade.client.SetTimeout(2 * time.Second)

	err = facade.Updates(context.Background(), metrics)
	require.NoError(t, err)

	// Empty metrics list should return nil without error
	err = facade.Updates(context.Background(), []*types.Metrics{})
	require.NoError(t, err)
}
