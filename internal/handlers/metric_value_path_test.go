package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	internalErrors "github.com/sbilibin2017/go-yandex-practicum/internal/errors"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestMetricGetPathHandler_ServeHTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		handler *MetricGetPathHandler
		w       *httptest.ResponseRecorder
		r       *http.Request
	)

	tests := []struct {
		name           string
		mType          string
		mName          string
		setup          func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:  "success",
			mType: "gauge",
			mName: "temperature",
			setup: func() {
				mockSvc := NewMockMetricGetPathService(ctrl)
				mockMetric := types.NewMetricFromAttributes("gauge", "temperature", "42.5")
				mockSvc.EXPECT().
					Get(gomock.Any(), types.MetricID{ID: "temperature", MType: "gauge"}).
					Return(mockMetric, nil)

				valFunc := func(mType, mName string) error {
					return nil
				}

				handler = NewMetricGetPathHandler(mockSvc, valFunc)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "42.5", // fixed: output of types.NewMetricString(...)
		},
		{
			name:  "validation error - metric name required",
			mType: "gauge",
			mName: "",
			setup: func() {
				valFunc := func(mType, mName string) error {
					return internalErrors.ErrMetricNameRequired
				}
				handler = NewMetricGetPathHandler(nil, valFunc)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:  "validation error - unsupported metric type",
			mType: "badtype",
			mName: "temperature",
			setup: func() {
				valFunc := func(mType, mName string) error {
					return internalErrors.ErrUnsupportedMetricType
				}
				handler = NewMetricGetPathHandler(nil, valFunc)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "service error",
			mType: "gauge",
			mName: "temperature",
			setup: func() {
				mockSvc := NewMockMetricGetPathService(ctrl)
				mockSvc.EXPECT().
					Get(gomock.Any(), types.MetricID{ID: "temperature", MType: "gauge"}).
					Return(nil, internalErrors.ErrMetricNameRequired)

				valFunc := func(mType, mName string) error {
					return nil
				}

				handler = NewMetricGetPathHandler(mockSvc, valFunc)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:  "metric not found",
			mType: "gauge",
			mName: "unknown",
			setup: func() {
				mockSvc := NewMockMetricGetPathService(ctrl)
				mockSvc.EXPECT().
					Get(gomock.Any(), types.MetricID{ID: "unknown", MType: "gauge"}).
					Return(nil, nil)

				valFunc := func(mType, mName string) error {
					return nil
				}

				handler = NewMetricGetPathHandler(mockSvc, valFunc)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			r = httptest.NewRequest(http.MethodGet, "/", nil)
			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("type", tt.mType)
			routeCtx.URLParams.Add("name", tt.mName)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

			w = httptest.NewRecorder()

			handler.ServeHTTP(w, r)

			assert.Equal(t, tt.expectedStatus, w.Result().StatusCode)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
				assert.Equal(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestHandleMetricGetPathError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		{
			name:           "ErrMetricNameRequired returns 404",
			err:            internalErrors.ErrMetricNameRequired,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "ErrUnsupportedMetricType returns 400",
			err:            internalErrors.ErrUnsupportedMetricType,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Unknown error returns 500",
			err:            assert.AnError,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Nil error returns 500",
			err:            nil,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handleMetricGetPathError(recorder, tt.err)
			assert.Equal(t, tt.expectedStatus, recorder.Result().StatusCode)
		})
	}
}
