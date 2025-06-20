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

func TestMetricGetBodyHandler_ServeHTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		handler *MetricGetBodyHandler
		w       *httptest.ResponseRecorder
		r       *http.Request
	)

	tests := []struct {
		name           string
		metricID       types.MetricID
		valErr         error
		serviceMetric  *types.Metrics
		serviceErr     error
		setup          func()
		expectedStatus int
		expectedBody   *types.Metrics
	}{
		{
			name:          "success",
			metricID:      types.MetricID{ID: "temperature"},
			valErr:        nil,
			serviceMetric: types.NewMetricFromAttributes("gauge", "temperature", "42.5"),
			serviceErr:    nil,
			setup: func() {
				mockSvc := NewMockMetricGetBodyService(ctrl)
				mockSvc.EXPECT().
					Get(gomock.Any(), gomock.Eq(types.MetricID{ID: "temperature"})).
					Return(types.NewMetricFromAttributes("gauge", "temperature", "42.5"), nil)

				valFunc := func(id types.MetricID) error {
					return nil
				}
				handler = NewMetricGetBodyHandler(mockSvc, valFunc)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   types.NewMetricFromAttributes("gauge", "temperature", "42.5"),
		},
		{
			name:     "json decode error",
			metricID: types.MetricID{},
			setup: func() {
				handler = NewMetricGetBodyHandler(nil, nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "validation error - ErrMetricIDRequired",
			metricID: types.MetricID{ID: ""},
			valErr:   internalErrors.ErrMetricIDRequired,
			setup: func() {
				valFunc := func(id types.MetricID) error {
					return internalErrors.ErrMetricIDRequired
				}
				handler = NewMetricGetBodyHandler(nil, valFunc)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "validation error - ErrUnsupportedMetricType",
			metricID: types.MetricID{ID: "bad-type"},
			valErr:   internalErrors.ErrUnsupportedMetricType,
			setup: func() {
				valFunc := func(id types.MetricID) error {
					return internalErrors.ErrUnsupportedMetricType
				}
				handler = NewMetricGetBodyHandler(nil, valFunc)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:          "service error",
			metricID:      types.MetricID{ID: "temperature"},
			serviceMetric: nil,
			serviceErr:    errors.New("service failure"),
			setup: func() {
				mockSvc := NewMockMetricGetBodyService(ctrl)
				mockSvc.EXPECT().
					Get(gomock.Any(), gomock.Eq(types.MetricID{ID: "temperature"})).
					Return(nil, errors.New("service failure"))

				valFunc := func(id types.MetricID) error {
					return nil
				}
				handler = NewMetricGetBodyHandler(mockSvc, valFunc)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:          "metric not found",
			metricID:      types.MetricID{ID: "notfound"},
			serviceMetric: nil,
			serviceErr:    nil,
			setup: func() {
				mockSvc := NewMockMetricGetBodyService(ctrl)
				mockSvc.EXPECT().
					Get(gomock.Any(), gomock.Eq(types.MetricID{ID: "notfound"})).
					Return(nil, nil)

				valFunc := func(id types.MetricID) error {
					return nil
				}
				handler = NewMetricGetBodyHandler(mockSvc, valFunc)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			var bodyBytes []byte
			if tt.name == "json decode error" {
				// invalid JSON
				bodyBytes = []byte(`{"invalid":}`)
			} else {
				bodyBytes, _ = json.Marshal(tt.metricID)
			}

			r = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			w = httptest.NewRecorder()

			handler.ServeHTTP(w, r)

			assert.Equal(t, tt.expectedStatus, w.Result().StatusCode)

			if tt.expectedStatus == http.StatusOK && tt.expectedBody != nil {
				var gotMetric types.Metrics
				err := json.NewDecoder(w.Body).Decode(&gotMetric)
				assert.NoError(t, err)
				assert.Equal(t, *tt.expectedBody, gotMetric)
			}
		})
	}
}
