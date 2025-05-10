package facades

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricFacade_Update_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/update/gauge/CPU/99.5"
		assert.Equal(t, expectedPath, r.URL.Path)
		assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	client := resty.New()
	mf := NewMetricFacade(*client, server.URL)
	metric := map[string]any{
		"type":  "gauge",
		"name":  "CPU",
		"value": "99.5",
	}
	err := mf.Update(context.Background(), metric)
	assert.NoError(t, err)
}

func TestMetricFacade_Update_ErrorFromServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer server.Close()
	client := resty.New()
	mf := NewMetricFacade(*client, server.URL)
	metric := map[string]any{
		"type":  "counter",
		"name":  "Requests",
		"value": "10",
	}
	err := mf.Update(context.Background(), metric)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error response from server")
}

func TestMetricFacade_Update_RequestError(t *testing.T) {
	client := resty.New()
	mf := NewMetricFacade(*client, "http://invalid-host")
	metric := map[string]any{
		"type":  "counter",
		"name":  "Errors",
		"value": "5",
	}
	err := mf.Update(context.Background(), metric)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send metric")
}
