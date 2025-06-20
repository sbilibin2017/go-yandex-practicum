package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	internalErrors "github.com/sbilibin2017/go-yandex-practicum/internal/errors"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestMetricUpdatesBodyHandler_ServeHTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name           string
		inputMetrics   []*types.Metrics
		inputRawBody   string
		valFunc        func(m *types.Metrics) error
		mockSvcSetup   func(m *MockMetricUpdatesBodyService)
		expectedStatus int
	}{
		{
			name: "successful update",
			inputMetrics: []*types.Metrics{
				types.NewMetricFromAttributes("gauge", "temp", "20.5"),
				types.NewMetricFromAttributes("counter", "requests", "5"),
			},
			valFunc: func(m *types.Metrics) error {
				return nil
			},
			mockSvcSetup: func(m *MockMetricUpdatesBodyService) {
				m.EXPECT().
					Updates(gomock.Any(), gomock.Any()).
					Return(nil, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid JSON",
			inputRawBody:   `{"id":"badjson"}`, // not a valid array of metrics
			valFunc:        func(m *types.Metrics) error { return nil },
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "validation failure on one metric",
			inputMetrics: []*types.Metrics{
				types.NewMetricFromAttributes("gauge", "good", "1"),
				// Manually create invalid metric to avoid nil return
				{ID: "invalid", MType: "badtype"},
			},
			valFunc: func(m *types.Metrics) error {
				if m == nil {
					return internalErrors.ErrUnsupportedMetricType
				}
				if m.ID == "invalid" {
					return internalErrors.ErrUnsupportedMetricType
				}
				return nil
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			inputMetrics: []*types.Metrics{
				types.NewMetricFromAttributes("counter", "req", "1"),
			},
			valFunc: func(m *types.Metrics) error { return nil },
			mockSvcSetup: func(m *MockMetricUpdatesBodyService) {
				m.EXPECT().
					Updates(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestBody []byte
			var err error

			if tt.inputRawBody != "" {
				requestBody = []byte(tt.inputRawBody)
			} else {
				requestBody, err = json.Marshal(tt.inputMetrics)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(requestBody))
			rec := httptest.NewRecorder()

			var mockSvc *MockMetricUpdatesBodyService
			if tt.mockSvcSetup != nil {
				mockSvc = NewMockMetricUpdatesBodyService(ctrl)
				tt.mockSvcSetup(mockSvc)
			}

			handler := NewMetricUpdatesBodyHandler(mockSvc, tt.valFunc)
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Result().StatusCode)
		})
	}
}

func TestHandleMetricUpdatesBodyError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		{
			name:           "metric ID required",
			err:            internalErrors.ErrMetricIDRequired,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "unsupported metric type",
			err:            internalErrors.ErrUnsupportedMetricType,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "counter value required",
			err:            internalErrors.ErrCounterValueRequired,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "gauge value required",
			err:            internalErrors.ErrGaugeValueRequired,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unexpected error",
			err:            assert.AnError, // generic error
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			handleMetricUpdatesBodyError(rec, tt.err)

			assert.Equal(t, tt.expectedStatus, rec.Result().StatusCode)
		})
	}
}
