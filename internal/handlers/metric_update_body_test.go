package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/handlers"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricUpdateBodyHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := handlers.NewMockMetricUpdateBodyService(ctrl)
	handler := handlers.NewMetricUpdateBodyHandler(mockService)
	tests := []struct {
		name           string
		metric         types.Metrics
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Valid Counter Metric",
			metric: types.Metrics{
				MetricID: types.MetricID{
					ID:   "counter1",
					Type: types.CounterMetricType,
				},
				Delta: ptrInt64(10),
			},
			mockSetup: func() {
				mockService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":"counter1","type":"counter","delta":10}`,
		},
		{
			name: "Valid Gauge Metric",
			metric: types.Metrics{
				MetricID: types.MetricID{
					ID:   "gauge1",
					Type: types.GaugeMetricType,
				},
				Value: ptrFloat64(10.5),
			},
			mockSetup: func() {
				mockService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":"gauge1","type":"gauge","value":10.5}`,
		},
		{
			name: "Missing Metric ID",
			metric: types.Metrics{
				MetricID: types.MetricID{
					ID:   "",
					Type: types.CounterMetricType,
				},
				Delta: ptrInt64(10),
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Metric id is required",
		},
		{
			name: "Invalid Metric Type",
			metric: types.Metrics{
				MetricID: types.MetricID{
					ID:   "invalid1",
					Type: "invalid",
				},
				Delta: ptrInt64(10),
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid metric type",
		},
		{
			name: "Metric Delta Missing for Counter",
			metric: types.Metrics{
				MetricID: types.MetricID{
					ID:   "counter1",
					Type: types.CounterMetricType,
				},
				Delta: nil,
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Metric delta is required for counter",
		},
		{
			name: "Metric Value Missing for Gauge",
			metric: types.Metrics{
				MetricID: types.MetricID{
					ID:   "gauge1",
					Type: types.GaugeMetricType,
				},
				Value: nil,
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Metric value is required for gauge",
		},
		{
			name: "Service Update Error",
			metric: types.Metrics{
				MetricID: types.MetricID{
					ID:   "counter1",
					Type: types.CounterMetricType,
				},
				Delta: ptrInt64(10),
			},
			mockSetup: func() {
				mockService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(types.ErrMetricInternal).Times(1)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Metric not updated",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			body, err := json.Marshal(tt.metric)
			assert.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/metrics/update", bytes.NewReader(body))
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, strings.TrimSpace(tt.expectedBody), strings.TrimSpace(rec.Body.String()))
		})
	}
}

func ptrInt64(i int64) *int64 {
	return &i
}

func ptrFloat64(f float64) *float64 {
	return &f
}
