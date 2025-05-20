package facades

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricFacade_Updates_EmptyMetrics(t *testing.T) {
	client := resty.New()
	mf := NewMetricFacade(client, "http://localhost", "X-Test-Header", "testkey")

	err := mf.Updates(context.Background(), []types.Metrics{})
	assert.NoError(t, err, "should not fail on empty metrics slice")
}

func TestMetricFacade_Updates_SuccessWithKey(t *testing.T) {
	metrics := []types.Metrics{
		{
			MetricID: types.MetricID{
				ID:   "metric1",
				Type: types.GaugeMetricType,
			},
			Value: ptrFloat64(123.45),
		},
		{
			MetricID: types.MetricID{
				ID:   "metric2",
				Type: types.CounterMetricType,
			},
			Delta: ptrInt64(10),
		},
	}

	bodyBytes, err := json.Marshal(metrics)
	assert.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/updates/", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Accept-Encoding"))

		// Проверяем HMAC-заголовок
		expectedHmac := computeHmac256(bodyBytes, []byte("testkey"))
		assert.Equal(t, expectedHmac, r.Header.Get("X-Signature"))

		// Проверяем тело запроса
		var received []types.Metrics
		err := json.NewDecoder(r.Body).Decode(&received)
		assert.NoError(t, err)
		assert.Equal(t, metrics, received)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := resty.New()
	mf := NewMetricFacade(client, server.URL, "X-Signature", "testkey")

	err = mf.Updates(context.Background(), metrics)
	assert.NoError(t, err)
}

func TestMetricFacade_Updates_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := resty.New()
	mf := NewMetricFacade(client, server.URL, "X-Signature", "testkey")

	metrics := []types.Metrics{
		{
			MetricID: types.MetricID{
				ID:   "m1",
				Type: types.GaugeMetricType,
			},
			Value: ptrFloat64(1.23),
		},
	}

	err := mf.Updates(context.Background(), metrics)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error response from server")
}

func TestMetricFacade_Updates_NetworkError(t *testing.T) {
	client := resty.New()
	mf := NewMetricFacade(client, "http://invalid.invalid", "X-Signature", "testkey")

	metrics := []types.Metrics{
		{
			MetricID: types.MetricID{
				ID:   "m1",
				Type: types.CounterMetricType,
			},
			Delta: ptrInt64(5),
		},
	}

	err := mf.Updates(context.Background(), metrics)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send metrics")
}

// --- Вспомогательные функции ---

func ptrFloat64(v float64) *float64 {
	return &v
}

func ptrInt64(v int64) *int64 {
	return &v
}

func computeHmac256(message, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(message)
	return hex.EncodeToString(h.Sum(nil))
}

func TestNewMetricFacade_AddsHTTPPrefix(t *testing.T) {
	client := resty.New()

	// Случай без префикса
	addr := "localhost:8080"
	mf := NewMetricFacade(client, addr, "X-Header", "secret")

	require.Equal(t, "http://"+addr, mf.serverAddress)

	// Случай с http://
	addr2 := "http://example.com"
	mf2 := NewMetricFacade(client, addr2, "X-Header", "secret")

	require.Equal(t, addr2, mf2.serverAddress)

	// Случай с https://
	addr3 := "https://secure.example.com"
	mf3 := NewMetricFacade(client, addr3, "X-Header", "secret")

	require.Equal(t, addr3, mf3.serverAddress)
}
