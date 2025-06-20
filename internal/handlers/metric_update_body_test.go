package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	internalErrors "github.com/sbilibin2017/go-yandex-practicum/internal/errors"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestMetricUpdateBodyHandler_ServeHTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		handler *MetricUpdateBodyHandler
		w       *httptest.ResponseRecorder
		r       *http.Request
	)

	tests := []struct {
		name           string
		metric         *types.Metrics
		valErr         error
		serviceErr     error
		setup          func()
		expectedStatus int
	}{
		{
			name:       "success",
			metric:     types.NewMetricFromAttributes("gauge", "temperature", "42.5"),
			valErr:     nil,
			serviceErr: nil,
			setup: func() {
				mockSvc := NewMockMetricUpdateBodyService(ctrl)
				mockSvc.EXPECT().
					Updates(gomock.Any(), gomock.Any()).
					Return([]*types.Metrics{}, nil)

				valFunc := func(metric *types.Metrics) error {
					return nil
				}

				handler = NewMetricUpdateBodyHandler(mockSvc, valFunc)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "json decode error",
			metric: nil, // will send invalid JSON
			setup: func() {
				// valFunc and service won't be called
				handler = NewMetricUpdateBodyHandler(nil, nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "validation error - ErrMetricIDRequired",
			metric: types.NewMetricFromAttributes("gauge", "", "42.5"),
			valErr: internalErrors.ErrMetricIDRequired,
			setup: func() {
				valFunc := func(metric *types.Metrics) error {
					return internalErrors.ErrMetricIDRequired
				}
				handler = NewMetricUpdateBodyHandler(nil, valFunc)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "validation error - ErrUnsupportedMetricType",
			metric: types.NewMetricFromAttributes("unsupported", "temperature", "42.5"),
			valErr: internalErrors.ErrUnsupportedMetricType,
			setup: func() {
				valFunc := func(metric *types.Metrics) error {
					return internalErrors.ErrUnsupportedMetricType
				}
				handler = NewMetricUpdateBodyHandler(nil, valFunc)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "service error",
			metric:     types.NewMetricFromAttributes("gauge", "temperature", "42.5"),
			valErr:     nil,
			serviceErr: internalErrors.ErrCounterValueRequired,
			setup: func() {
				mockSvc := NewMockMetricUpdateBodyService(ctrl)
				mockSvc.EXPECT().
					Updates(gomock.Any(), gomock.Any()).
					Return(nil, internalErrors.ErrCounterValueRequired)

				valFunc := func(metric *types.Metrics) error {
					return nil
				}

				handler = NewMetricUpdateBodyHandler(mockSvc, valFunc)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "unknown service error",
			metric:     types.NewMetricFromAttributes("gauge", "temperature", "42.5"),
			valErr:     nil,
			serviceErr: errors.New("unknown error"),
			setup: func() {
				mockSvc := NewMockMetricUpdateBodyService(ctrl)
				mockSvc.EXPECT().
					Updates(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("unknown error"))

				valFunc := func(metric *types.Metrics) error {
					return nil
				}

				handler = NewMetricUpdateBodyHandler(mockSvc, valFunc)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable

		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			var bodyBytes []byte
			if tt.metric != nil {
				bodyBytes, _ = json.Marshal(tt.metric)
			} else {
				bodyBytes = []byte("invalid-json")
			}

			r = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			w = httptest.NewRecorder()

			handler.ServeHTTP(w, r)

			assert.Equal(t, tt.expectedStatus, w.Result().StatusCode)
		})
	}
}
