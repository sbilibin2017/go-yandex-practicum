package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricUpdatesBodyHandler(t *testing.T) {
	type testCase struct {
		name           string
		requestBody    []types.Metrics
		mockSetup      func(svc *MockMetricUpdatesBodyService)
		expectedStatus int
		expectedBody   string
	}

	validGauge := types.Metrics{
		MetricID: types.MetricID{
			ID:   "temperature",
			Type: types.GaugeMetricType,
		},
		Value: float64Ptr(36.6),
	}

	validCounter := types.Metrics{
		MetricID: types.MetricID{
			ID:   "requests",
			Type: types.CounterMetricType,
		},
		Delta: int64Ptr(1),
	}

	tests := []testCase{
		{
			name:        "valid gauge and counter metrics",
			requestBody: []types.Metrics{validGauge, validCounter},
			mockSetup: func(svc *MockMetricUpdatesBodyService) {
				svc.EXPECT().
					Updates(gomock.Any(), gomock.Any()).
					Return([]types.Metrics{validGauge, validCounter}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   mustMarshalJSON([]types.Metrics{validGauge, validCounter}),
		},
		{
			name: "invalid metric - missing ID",
			requestBody: []types.Metrics{{
				MetricID: types.MetricID{
					Type: types.CounterMetricType,
				},
				Delta: int64Ptr(1),
			}},
			mockSetup:      func(svc *MockMetricUpdatesBodyService) {},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Metric id is required",
		},
		{
			name: "invalid type",
			requestBody: []types.Metrics{{
				MetricID: types.MetricID{
					ID:   "badmetric",
					Type: "invalid",
				},
				Value: float64Ptr(10.5),
			}},
			mockSetup:      func(svc *MockMetricUpdatesBodyService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid metric type",
		},
		{
			name: "missing delta in counter",
			requestBody: []types.Metrics{{
				MetricID: types.MetricID{
					ID:   "count",
					Type: types.CounterMetricType,
				},
			}},
			mockSetup:      func(svc *MockMetricUpdatesBodyService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Metric delta is required for counter",
		},
		{
			name:        "service returns error",
			requestBody: []types.Metrics{validCounter},
			mockSetup: func(svc *MockMetricUpdatesBodyService) {
				svc.EXPECT().
					Updates(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Metric not updated",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := NewMockMetricUpdatesBodyService(ctrl)
			tc.mockSetup(mockSvc)

			handler := NewMetricUpdatesBodyHandler(mockSvc)

			bodyBytes, err := json.Marshal(tc.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			resp := rec.Result()
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			assert.Equal(t, tc.expectedBody, string(bytes.TrimSpace(body)))
		})
	}
}

// Helpers

func int64Ptr(i int64) *int64 {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func mustMarshalJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}
