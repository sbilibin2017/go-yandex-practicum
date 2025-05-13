package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/handlers"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestNewMetricGetBodyHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := handlers.NewMockMetricGetBodyService(ctrl)
	handler := handlers.NewMetricGetBodyHandler(mockService)
	ptrInt64 := func(v int64) *int64 {
		return &v
	}
	t.Run("success - counter", func(t *testing.T) {
		metricID := types.MetricID{
			ID:   "test_counter",
			Type: types.CounterMetricType,
		}
		metric := &types.Metrics{
			MetricID: metricID,
			Delta:    ptrInt64(42),
		}
		body, _ := json.Marshal(metricID)
		req := httptest.NewRequest(http.MethodPost, "/get/", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()
		mockService.EXPECT().Get(gomock.Any(), metricID).Return(metric, nil)
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var resp types.Metrics
		err := json.NewDecoder(rec.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, metric.MetricID, resp.MetricID)
		assert.Equal(t, *metric.Delta, *resp.Delta)
	})
	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/get/", bytes.NewBufferString("{invalid"))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid JSON body")
	})
	t.Run("missing ID", func(t *testing.T) {
		metricID := types.MetricID{
			ID:   "",
			Type: types.CounterMetricType,
		}
		body, _ := json.Marshal(metricID)
		req := httptest.NewRequest(http.MethodPost, "/get/", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Metric ID is required")
	})
	t.Run("invalid metric type", func(t *testing.T) {
		metricID := types.MetricID{
			ID:   "bad_type",
			Type: "unknown",
		}
		body, _ := json.Marshal(metricID)
		req := httptest.NewRequest(http.MethodPost, "/get/", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid metric type")
	})
	t.Run("metric not found", func(t *testing.T) {
		metricID := types.MetricID{
			ID:   "not_found",
			Type: types.GaugeMetricType,
		}
		body, _ := json.Marshal(metricID)
		req := httptest.NewRequest(http.MethodPost, "/get/", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()
		mockService.EXPECT().Get(gomock.Any(), metricID).Return(nil, types.ErrMetricNotFound)
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "Metric not found")
	})
	t.Run("service error", func(t *testing.T) {
		metricID := types.MetricID{
			ID:   "internal_error",
			Type: types.GaugeMetricType,
		}
		body, _ := json.Marshal(metricID)
		req := httptest.NewRequest(http.MethodPost, "/get/", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()
		mockService.EXPECT().Get(gomock.Any(), metricID).Return(nil, errors.New("fail"))
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Internal server error")
	})
	t.Run("json encode error", func(t *testing.T) {
		metricID := types.MetricID{
			ID:   "encode_fail",
			Type: types.GaugeMetricType,
		}
		body, _ := json.Marshal(metricID)
		req := httptest.NewRequest(http.MethodPost, "/get/", bytes.NewBuffer(body))
		brokenWriter := &errorValueBodyWriter{}
		mockService.EXPECT().Get(gomock.Any(), metricID).Return(&types.Metrics{}, nil)
		handler.ServeHTTP(brokenWriter, req)
		assert.Equal(t, http.StatusInternalServerError, brokenWriter.statusCode)
		assert.Contains(t, brokenWriter.body, "Failed to encode response")
	})
}

type errorValueBodyWriter struct {
	header     http.Header
	statusCode int
	body       string
}

func (w *errorValueBodyWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *errorValueBodyWriter) Write(b []byte) (int, error) {
	w.body = string(b)
	return 0, errors.New("write error")
}

func (w *errorValueBodyWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}
