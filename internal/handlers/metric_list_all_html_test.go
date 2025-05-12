package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestMetricListAllHTMLHandler(t *testing.T) {
	tests := []struct {
		name           string
		mockMetrics    []types.Metrics
		mockErr        error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success with gauge and counter",
			mockMetrics: []types.Metrics{
				*types.NewMetrics("temperature", string(types.GaugeMetricType), nil, floatPtr(21.5)),
				*types.NewMetrics("requests", string(types.CounterMetricType), intPtr(42), nil),
			},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
			expectedBody:   "<li>temperature: 21.5</li>",
		},
		{
			name:           "empty list",
			mockMetrics:    []types.Metrics{},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
			expectedBody:   "<ul>\n</ul>",
		},
		{
			name:           "internal error",
			mockMetrics:    nil,
			mockErr:        errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := NewMockMetricListAllHTMLService(ctrl)
			mockService.EXPECT().
				ListAll(gomock.Any()).
				Return(tt.mockMetrics, tt.mockErr)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()

			handler := MetricListAllHTMLHandler(mockService)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			} else {
				assert.Equal(t, tt.expectedBody, rr.Body.String())
			}
		})
	}
}

func floatPtr(v float64) *float64 {
	return &v
}

func intPtr(v int64) *int64 {
	return &v
}
