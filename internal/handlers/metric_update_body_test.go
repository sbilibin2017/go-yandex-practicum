package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestNewMetricUpdateBodyHandler(t *testing.T) {
	type testCase struct {
		name           string
		body           interface{}
		mockService    func(m *MockMetricUpdateBodyService)
		expectedStatus int
		expectedBody   string
	}

	counterDelta := int64(5)
	gaugeValue := 3.14
	successResp := []types.Metrics{
		{
			MetricID: types.MetricID{ID: "cpu", Type: types.CounterMetricType},
			Delta:    &counterDelta,
		},
	}

	tests := []testCase{
		{
			name: "valid counter metric",
			body: types.Metrics{
				MetricID: types.MetricID{ID: "cpu", Type: types.CounterMetricType},
				Delta:    &counterDelta,
			},
			mockService: func(m *MockMetricUpdateBodyService) {
				m.EXPECT().
					Updates(gomock.Any(), gomock.Any()).
					Return(successResp, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "valid gauge metric",
			body: types.Metrics{
				MetricID: types.MetricID{ID: "memory", Type: types.GaugeMetricType},
				Value:    &gaugeValue,
			},
			mockService: func(m *MockMetricUpdateBodyService) {
				m.EXPECT().
					Updates(gomock.Any(), gomock.Any()).
					Return([]types.Metrics{
						{
							MetricID: types.MetricID{ID: "memory", Type: types.GaugeMetricType},
							Value:    &gaugeValue,
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid JSON",
			body:           `{"invalid`,
			mockService:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON body\n",
		},
		{
			name: "missing metric id",
			body: types.Metrics{
				MetricID: types.MetricID{ID: "", Type: types.CounterMetricType},
				Delta:    &counterDelta,
			},
			mockService:    nil,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Metric id is required\n",
		},
		{
			name: "missing delta in counter",
			body: types.Metrics{
				MetricID: types.MetricID{ID: "cpu", Type: types.CounterMetricType},
			},
			mockService:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Metric delta is required for counter\n",
		},
		{
			name: "missing value in gauge",
			body: types.Metrics{
				MetricID: types.MetricID{ID: "mem", Type: types.GaugeMetricType},
			},
			mockService:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Metric value is required for gauge\n",
		},
		{
			name: "invalid metric type",
			body: types.Metrics{
				MetricID: types.MetricID{ID: "unknown", Type: "other"},
			},
			mockService:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid metric type\n",
		},
		{
			name: "service error",
			body: types.Metrics{
				MetricID: types.MetricID{ID: "cpu", Type: types.CounterMetricType},
				Delta:    &counterDelta,
			},
			mockService: func(m *MockMetricUpdateBodyService) {
				m.EXPECT().
					Updates(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("some error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Metric not updated\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := NewMockMetricUpdateBodyService(ctrl)
			if tc.mockService != nil {
				tc.mockService(mockService)
			}

			handler := NewMetricUpdateBodyHandler(mockService)

			var reqBody []byte
			switch b := tc.body.(type) {
			case string:
				reqBody = []byte(b)
			default:
				var err error
				reqBody, err = json.Marshal(b)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBuffer(reqBody))
			w := httptest.NewRecorder()

			handler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			if tc.expectedBody != "" {
				buf := new(bytes.Buffer)
				_, _ = buf.ReadFrom(resp.Body)
				assert.Equal(t, tc.expectedBody, buf.String())
			}
		})
	}
}
