package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/go-yandex-practicum/internal/handlers"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestMetricGetPathHandler_WithRouter(t *testing.T) {
	tests := []struct {
		name           string
		mtype          string
		metricName     string
		mockReturn     *types.Metrics
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Gauge metric success",
			mtype:          "gauge",
			metricName:     "temperature",
			mockReturn:     &types.Metrics{MetricID: types.MetricID{ID: "temperature", Type: types.GaugeMetricType}, Value: floatPtr(23.5)},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Counter metric success",
			mtype:          "counter",
			metricName:     "requests",
			mockReturn:     &types.Metrics{MetricID: types.MetricID{ID: "requests", Type: types.CounterMetricType}, Delta: int64Ptr(101)},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Metric not found",
			mtype:          "gauge",
			metricName:     "unknown",
			mockError:      types.ErrMetricNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Internal error",
			mtype:          "counter",
			metricName:     "failmetric",
			mockError:      errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Invalid metric type",
			mtype:          "invalid",
			metricName:     "bad",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing metric name",
			mtype:          string(types.CounterMetricType),
			metricName:     "",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := handlers.NewMockMetricGetPathService(ctrl)

			if tt.mockReturn != nil || tt.mockError != nil {
				mockService.EXPECT().
					Get(gomock.Any(), types.MetricID{ID: tt.metricName, Type: types.MetricType(tt.mtype)}).
					Return(tt.mockReturn, tt.mockError).
					Times(1)
			}

			handler := handlers.NewMetricGetPathHandler(mockService)

			router := chi.NewRouter()
			router.Get("/value/{type}/{name}", handler)

			target := "/value/" + tt.mtype + "/" + tt.metricName
			req, err := http.NewRequest("GET", target, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

		})
	}
}

func floatPtr(v float64) *float64 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}
