package handlers

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricUpdatePathHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := NewMockMetricUpdatePathService(ctrl)
	handler := MetricUpdatePathHandler(mockService)

	tests := []struct {
		name  string
		url   string
		setup func()
		want  struct {
			statusCode int
			response   string
		}
	}{
		{
			name: "Valid counter metric update",
			url:  "/update/counter/some_metric/10",
			setup: func() {
				delta, err := strconv.ParseInt("10", 10, 64)
				require.NoError(t, err)
				metrics := []types.Metrics{
					{
						MetricID: types.MetricID{
							ID:   "some_metric",
							Type: types.CounterMetricType,
						},
						Delta: &delta,
					},
				}
				mockService.EXPECT().Update(gomock.Any(), gomock.Eq(metrics)).Return(nil).Times(1)
			},
			want: struct {
				statusCode int
				response   string
			}{
				statusCode: http.StatusOK,
				response:   "Metric updated successfully",
			},
		},
		{
			name: "Valid gauge metric update",
			url:  "/update/gauge/some_metric/10.5",
			setup: func() {
				val, err := strconv.ParseFloat("10.5", 64)
				require.NoError(t, err)
				metrics := []types.Metrics{
					{
						MetricID: types.MetricID{
							ID:   "some_metric",
							Type: types.GaugeMetricType,
						},
						Value: &val,
					},
				}
				mockService.EXPECT().Update(gomock.Any(), gomock.Eq(metrics)).Return(nil).Times(1)
			},
			want: struct {
				statusCode int
				response   string
			}{
				statusCode: http.StatusOK,
				response:   "Metric updated successfully",
			},
		},
		{
			name:  "Invalid metric type",
			url:   "/update/invalid/some_metric/10",
			setup: func() {},
			want: struct {
				statusCode int
				response   string
			}{
				statusCode: http.StatusBadRequest,
				response:   "Invalid metric type",
			},
		},
		{
			name:  "Invalid metric value for counter",
			url:   "/update/counter/some_metric/invalid_value",
			setup: func() {},
			want: struct {
				statusCode int
				response   string
			}{
				statusCode: http.StatusBadRequest,
				response:   "Invalid metric value for counter",
			},
		},
		{
			name: "Service update error",
			url:  "/update/counter/some_metric/10",
			setup: func() {
				delta := int64(10)
				metrics := []types.Metrics{
					{
						MetricID: types.MetricID{
							ID:   "some_metric",
							Type: types.CounterMetricType,
						},
						Delta: &delta,
					},
				}
				mockService.EXPECT().Update(gomock.Any(), gomock.Eq(metrics)).Return(types.ErrMetricIsNotUpdated).Times(1)
			},
			want: struct {
				statusCode int
				response   string
			}{
				statusCode: http.StatusInternalServerError,
				response:   "Metric not updated",
			},
		},
		{
			name:  "Invalid metric value for counter (non-numeric)",
			url:   "/update/counter/some_metric/invalid_value",
			setup: func() {},
			want: struct {
				statusCode int
				response   string
			}{
				statusCode: http.StatusBadRequest,
				response:   "Invalid metric value for counter",
			},
		},
		{
			name:  "Invalid metric value for gauge",
			url:   "/update/gauge/some_metric/invalid_value",
			setup: func() {},
			want: struct {
				statusCode int
				response   string
			}{
				statusCode: http.StatusBadRequest,
				response:   "Invalid metric value for gauge",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			req, err := http.NewRequest(http.MethodPost, tt.url, nil)
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			assert.Equal(t, tt.want.statusCode, rr.Code)
			assert.Equal(t, tt.want.response, strings.TrimSpace(rr.Body.String()))
		})
	}
}
