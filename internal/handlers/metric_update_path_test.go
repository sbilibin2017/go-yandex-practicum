package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricUpdatePathHandler(t *testing.T) {
	type testCase struct {
		name           string
		urlPath        string
		mockService    func(m *MockMetricUpdatePathService)
		expectedStatus int
		expectedBody   string
	}

	tests := []testCase{
		{
			name:    "valid counter metric",
			urlPath: "/update/counter/ops/100",
			mockService: func(m *MockMetricUpdatePathService) {
				delta := int64(100)
				m.EXPECT().
					Updates(gomock.Any(), []types.Metrics{
						{
							MetricID: types.MetricID{ID: "ops", Type: types.CounterMetricType},
							Delta:    &delta,
						},
					}).
					Return(nil, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Metric updated successfully",
		},
		{
			name:    "valid gauge metric",
			urlPath: "/update/gauge/load/3.14",
			mockService: func(m *MockMetricUpdatePathService) {
				value := 3.14
				m.EXPECT().
					Updates(gomock.Any(), []types.Metrics{
						{
							MetricID: types.MetricID{ID: "load", Type: types.GaugeMetricType},
							Value:    &value,
						},
					}).
					Return(nil, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Metric updated successfully",
		},
		{
			name:           "invalid metric type",
			urlPath:        "/update/unknown/test/42",
			mockService:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid metric type\n",
		},
		{
			name:           "invalid counter value",
			urlPath:        "/update/counter/test/abc",
			mockService:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid metric value for counter\n",
		},
		{
			name:           "invalid gauge value",
			urlPath:        "/update/gauge/test/abc",
			mockService:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid metric value for gauge\n",
		},
		{
			name:    "internal error",
			urlPath: "/update/counter/test/1",
			mockService: func(m *MockMetricUpdatePathService) {
				delta := int64(1)
				m.EXPECT().
					Updates(gomock.Any(), []types.Metrics{
						{
							MetricID: types.MetricID{ID: "test", Type: types.CounterMetricType},
							Delta:    &delta,
						},
					}).
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Metric not updated\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := NewMockMetricUpdatePathService(ctrl)
			if tc.mockService != nil {
				tc.mockService(mockService)
			}

			router := chi.NewRouter()
			router.Post("/update/{type}/{name}/{value}", NewMetricUpdatePathHandler(mockService))

			req := httptest.NewRequest(http.MethodPost, tc.urlPath, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

		})
	}
}
