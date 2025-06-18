package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestNewMetricUpdatePathHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockMetricUpdatePathService(ctrl)
	handler := NewMetricUpdatePathHandler(mockSvc)

	tests := []struct {
		name           string
		query          string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:           "Missing name param",
			query:          "type=counter&value=10",
			mockSetup:      func() {},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Missing value param",
			query:          "name=metric1&type=counter",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid counter value",
			query:          "name=metric1&type=counter&value=notint",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid gauge value",
			query:          "name=metric1&type=gauge&value=notfloat",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid metric type",
			query:          "name=metric1&type=invalid&value=123",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "Service returns error",
			query: "name=metric1&type=counter&value=123",
			mockSetup: func() {
				mockSvc.EXPECT().
					Updates(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:  "Successful update counter",
			query: "name=metric1&type=counter&value=123",
			mockSetup: func() {
				mockSvc.EXPECT().
					Updates(gomock.Any(), []*types.Metrics{
						{
							ID:    "metric1",
							Type:  types.Counter,
							Delta: func() *int64 { v := int64(123); return &v }(),
						},
					}).
					Return(nil, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "Successful update gauge",
			query: "name=metric1&type=gauge&value=123.45",
			mockSetup: func() {
				mockSvc.EXPECT().
					Updates(gomock.Any(), []*types.Metrics{
						{
							ID:    "metric1",
							Type:  types.Gauge,
							Value: func() *float64 { v := 123.45; return &v }(),
						},
					}).
					Return(nil, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodPost, "/update?"+tt.query, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.expectedStatus, res.StatusCode)
		})
	}
}
