package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/handlers"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricGetPathHandler(t *testing.T) {
	int64Ptr := func(i int64) *int64 { return &i }
	float64Ptr := func(f float64) *float64 { return &f }

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := handlers.NewMockMetricGetPathService(ctrl)
	handler := handlers.NewMetricGetPathHandler(mockSvc)

	tests := []struct {
		name           string
		metricType     string
		metricID       string
		mockReturn     *types.Metrics
		mockErr        error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid Counter metric with delta",
			metricType:     types.Counter,
			metricID:       "counter1",
			mockReturn:     &types.Metrics{ID: "counter1", Type: types.Counter, Delta: int64Ptr(123)},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
			expectedBody:   "123",
		},
		{
			name:           "Valid Gauge metric with value",
			metricType:     types.Gauge,
			metricID:       "gauge1",
			mockReturn:     &types.Metrics{ID: "gauge1", Type: types.Gauge, Value: float64Ptr(45.67)},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
			expectedBody:   "45.67",
		},
		{
			name:           "Missing metric id returns 404",
			metricType:     types.Gauge,
			metricID:       "",
			mockReturn:     nil,
			mockErr:        nil,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
		},
		{
			name:           "Invalid metric type returns 400",
			metricType:     "invalid-type",
			metricID:       "someid",
			mockReturn:     nil,
			mockErr:        nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:           "Service returns error returns 500",
			metricType:     types.Counter,
			metricID:       "counter1",
			mockReturn:     nil,
			mockErr:        errors.New("some error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "",
		},
		{
			name:           "Service returns nil metric returns 404",
			metricType:     types.Counter,
			metricID:       "notfound",
			mockReturn:     nil,
			mockErr:        nil,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.metricID != "" && (tt.metricType == types.Counter || tt.metricType == types.Gauge) {
				mockSvc.EXPECT().
					Get(gomock.Any(), types.MetricID{ID: tt.metricID, Type: tt.metricType}).
					Return(tt.mockReturn, tt.mockErr).
					Times(1)
			} else {
				// No call expected if inputs invalid
				mockSvc.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Times(0)
			}

			req := httptest.NewRequest("GET", "/metric/"+tt.metricType+"/"+tt.metricID, nil)
			rctx := chi.NewRouteContext()
			if tt.metricType != "" {
				rctx.URLParams.Add("type", tt.metricType)
			}
			if tt.metricID != "" {
				rctx.URLParams.Add("id", tt.metricID)
			}
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, strings.TrimSpace(tt.expectedBody), strings.TrimSpace(rr.Body.String()))
		})
	}
}
