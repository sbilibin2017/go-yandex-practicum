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

func TestNewMetricUpdateBodyHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := handlers.NewMockMetricUpdateBodyService(ctrl)
	handler := handlers.NewMetricUpdateBodyHandler(mockService)
	ptrInt64 := func(v int64) *int64 {
		return &v
	}
	ptrFloat64 := func(v float64) *float64 {
		return &v
	}
	t.Run("success - counter", func(t *testing.T) {
		metric := types.Metrics{
			MetricID: types.MetricID{
				ID:   "test_counter",
				Type: types.CounterMetricType,
			},
			Delta: ptrInt64(42),
		}
		body, _ := json.Marshal(metric)
		req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()
		mockService.EXPECT().Update(gomock.Any(), []types.Metrics{metric}).Return(nil)
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var resp types.Metrics
		err := json.NewDecoder(rec.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, metric.ID, resp.ID)
		assert.Equal(t, *metric.Delta, *resp.Delta)
	})
	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBufferString("{invalid"))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid JSON body")
	})
	t.Run("missing ID", func(t *testing.T) {
		metric := types.Metrics{
			MetricID: types.MetricID{
				ID:   "",
				Type: types.CounterMetricType,
			},
			Delta: ptrInt64(5),
		}
		body, _ := json.Marshal(metric)
		req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "Metric id is required")
	})
	t.Run("invalid metric type", func(t *testing.T) {
		metric := types.Metrics{
			MetricID: types.MetricID{
				ID:   "bad_type",
				Type: "unknown",
			},
		}
		body, _ := json.Marshal(metric)
		req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid metric type")
	})
	t.Run("missing delta for counter", func(t *testing.T) {
		metric := types.Metrics{
			MetricID: types.MetricID{
				ID:   "missing_delta",
				Type: types.CounterMetricType,
			},
			Delta: nil,
		}
		body, _ := json.Marshal(metric)
		req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Metric delta is required")
	})
	t.Run("missing value for gauge", func(t *testing.T) {
		metric := types.Metrics{
			MetricID: types.MetricID{
				ID:   "missing_value",
				Type: types.GaugeMetricType,
			},
			Value: nil,
		}
		body, _ := json.Marshal(metric)
		req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Metric value is required")
	})
	t.Run("service error", func(t *testing.T) {
		metric := types.Metrics{
			MetricID: types.MetricID{
				ID:   "internal_error",
				Type: types.GaugeMetricType,
			},
			Value: ptrFloat64(42.42),
		}
		body, _ := json.Marshal(metric)
		req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer(body))
		rec := httptest.NewRecorder()
		mockService.EXPECT().Update(gomock.Any(), []types.Metrics{metric}).Return(errors.New("fail"))
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Metric not updated")
	})
	t.Run("json encode error", func(t *testing.T) {
		metric := types.Metrics{
			MetricID: types.MetricID{
				ID:   "encode_fail",
				Type: types.GaugeMetricType,
			},
			Value: ptrFloat64(99.99),
		}
		body, _ := json.Marshal(metric)
		req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer(body))
		brokenWriter := &errorWriter{}
		mockService.EXPECT().Update(gomock.Any(), []types.Metrics{metric}).Return(nil)
		handler.ServeHTTP(brokenWriter, req)
		assert.Equal(t, http.StatusInternalServerError, brokenWriter.statusCode)
		assert.Contains(t, brokenWriter.body, "Failed to encode response")
	})
}

type errorWriter struct {
	header     http.Header
	statusCode int
	body       string
}

func (w *errorWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *errorWriter) Write(b []byte) (int, error) {
	w.body = string(b)
	return 0, errors.New("write error")
}

func (w *errorWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}
