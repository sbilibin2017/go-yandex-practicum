package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	internalErrors "github.com/sbilibin2017/go-yandex-practicum/internal/errors"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestMetricUpdatePathHandler_ServeHTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		handler *MetricUpdatePathHandler
		w       *httptest.ResponseRecorder
		r       *http.Request
	)

	tests := []struct {
		name           string
		mType          string
		mName          string
		mValue         string
		valErr         error
		setup          func()
		expectedStatus int
	}{
		{
			name:   "success",
			mType:  "gauge",
			mName:  "temperature",
			mValue: "42.5",
			valErr: nil,
			setup: func() {
				mockSvc := NewMockMetricUpdatePathService(ctrl)
				mockSvc.EXPECT().
					Updates(gomock.Any(), gomock.Any()).
					Return([]*types.Metrics{types.NewMetricFromAttributes("gauge", "temperature", "42.5")}, nil)

				valFunc := func(mType, mName, mValue string) error {
					return nil
				}

				handler = NewMetricUpdatePathHandler(mockSvc, valFunc)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "validation error - metric name required",
			mType:  "gauge",
			mName:  "",
			mValue: "42.5",
			valErr: internalErrors.ErrMetricNameRequired,
			setup: func() {
				valFunc := func(mType, mName, mValue string) error {
					return internalErrors.ErrMetricNameRequired
				}
				// mock service won't be called on validation error, so nil is fine here
				handler = NewMetricUpdatePathHandler(nil, valFunc)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "service error",
			mType:  "gauge",
			mName:  "temperature",
			mValue: "42.5",
			valErr: nil,
			setup: func() {
				mockSvc := NewMockMetricUpdatePathService(ctrl)
				mockSvc.EXPECT().
					Updates(gomock.Any(), gomock.Any()).
					Return(nil, internalErrors.ErrInvalidCounterValue)

				valFunc := func(mType, mName, mValue string) error {
					return nil
				}

				handler = NewMetricUpdatePathHandler(mockSvc, valFunc)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "unknown service error",
			mType:  "gauge",
			mName:  "temperature",
			mValue: "42.5",
			valErr: nil,
			setup: func() {
				mockSvc := NewMockMetricUpdatePathService(ctrl)
				mockSvc.EXPECT().
					Updates(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("unknown error"))

				valFunc := func(mType, mName, mValue string) error {
					return nil
				}

				handler = NewMetricUpdatePathHandler(mockSvc, valFunc)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable for parallel tests if needed

		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			// Setup request with chi URL params
			r = httptest.NewRequest(http.MethodPost, "/", nil)
			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("type", tt.mType)
			routeCtx.URLParams.Add("name", tt.mName)
			routeCtx.URLParams.Add("value", tt.mValue)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

			w = httptest.NewRecorder()

			handler.ServeHTTP(w, r)
			assert.Equal(t, tt.expectedStatus, w.Result().StatusCode)
		})
	}
}
