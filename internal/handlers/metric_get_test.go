package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricGetPathHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ptrInt64 := func(i int64) *int64 { return &i }
	ptrFloat64 := func(f float64) *float64 { return &f }

	mockGetter := NewMockMetricGetter(ctrl)
	handler := NewMetricGetPathHandler(WithMetricGetterPath(mockGetter))

	r := chi.NewRouter()
	handler.RegisterRoute(r)

	tests := []struct {
		name         string
		method       string
		url          string
		mockExpect   func()
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Missing metric name (path parameter missing)",
			method:       http.MethodGet,
			url:          "/value/counter", // no name param at all
			mockExpect:   func() {},
			expectedCode: http.StatusNotFound,
		},
		{
			name:   "Valid counter metric",
			method: http.MethodGet,
			url:    "/value/counter/myCounter",
			mockExpect: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), types.MetricID{ID: "myCounter", Type: types.Counter}).
					Return(&types.Metrics{ID: "myCounter", Type: types.Counter, Delta: ptrInt64(42)}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: "42",
		},
		{
			name:   "Valid gauge metric",
			method: http.MethodGet,
			url:    "/value/gauge/myGauge",
			mockExpect: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), types.MetricID{ID: "myGauge", Type: types.Gauge}).
					Return(&types.Metrics{ID: "myGauge", Type: types.Gauge, Value: ptrFloat64(3.14)}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: "3.14",
		},
		{
			name:         "Missing metric name",
			method:       http.MethodGet,
			url:          "/value/counter/", // triggers name == "" case
			mockExpect:   func() {},
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "Invalid metric type",
			method:       http.MethodGet,
			url:          "/value/unknown/myMetric",
			mockExpect:   func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:   "Metric not found",
			method: http.MethodGet,
			url:    "/value/counter/missingMetric",
			mockExpect: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), types.MetricID{ID: "missingMetric", Type: types.Counter}).
					Return(nil, nil)
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name:   "Getter returns error",
			method: http.MethodGet,
			url:    "/value/gauge/errorMetric",
			mockExpect: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), types.MetricID{ID: "errorMetric", Type: types.Gauge}).
					Return(nil, context.DeadlineExceeded)
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:   "Counter metric with nil Delta",
			method: http.MethodGet,
			url:    "/value/counter/nilDelta",
			mockExpect: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), types.MetricID{ID: "nilDelta", Type: types.Counter}).
					Return(&types.Metrics{ID: "nilDelta", Type: types.Counter, Delta: nil}, nil)
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name:   "Gauge metric with nil Value",
			method: http.MethodGet,
			url:    "/value/gauge/nilValue",
			mockExpect: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), types.MetricID{ID: "nilValue", Type: types.Gauge}).
					Return(&types.Metrics{ID: "nilValue", Type: types.Gauge, Value: nil}, nil)
			},
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockExpect()

			req := httptest.NewRequest(tt.method, tt.url, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestMetricGetBodyHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ptrInt64 := func(i int64) *int64 { return &i }
	ptrFloat64 := func(f float64) *float64 { return &f }

	mockGetter := NewMockMetricGetter(ctrl)
	handler := NewMetricGetBodyHandler(WithMetricGetterBody(mockGetter))

	r := chi.NewRouter()
	handler.RegisterRoute(r)

	tests := []struct {
		name         string
		method       string
		url          string
		requestBody  string
		mockExpect   func()
		expectedCode int
		expectedBody string
	}{
		{
			name:        "Valid counter metric",
			method:      http.MethodPost,
			url:         "/value/",
			requestBody: `{"id":"myCounter","type":"counter"}`,
			mockExpect: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), types.MetricID{ID: "myCounter", Type: types.Counter}).
					Return(&types.Metrics{ID: "myCounter", Type: types.Counter, Delta: ptrInt64(123)}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"myCounter","type":"counter","delta":123}`,
		},
		{
			name:        "Valid gauge metric",
			method:      http.MethodPost,
			url:         "/value/",
			requestBody: `{"id":"myGauge","type":"gauge"}`,
			mockExpect: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), types.MetricID{ID: "myGauge", Type: types.Gauge}).
					Return(&types.Metrics{ID: "myGauge", Type: types.Gauge, Value: ptrFloat64(9.87)}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"myGauge","type":"gauge","value":9.87}`,
		},
		{
			name:         "Invalid JSON body",
			method:       http.MethodPost,
			url:          "/value/",
			requestBody:  `{"id": "missing quote}`, // invalid JSON
			mockExpect:   func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Missing metric ID",
			method:       http.MethodPost,
			url:          "/value/",
			requestBody:  `{"type":"counter"}`,
			mockExpect:   func() {},
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "Invalid metric type",
			method:       http.MethodPost,
			url:          "/value/",
			requestBody:  `{"id":"myMetric","type":"unknown"}`,
			mockExpect:   func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:        "Getter returns error",
			method:      http.MethodPost,
			url:         "/value/",
			requestBody: `{"id":"errorMetric","type":"gauge"}`,
			mockExpect: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), types.MetricID{ID: "errorMetric", Type: types.Gauge}).
					Return(nil, context.DeadlineExceeded)
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:        "Metric not found",
			method:      http.MethodPost,
			url:         "/value/",
			requestBody: `{"id":"missingMetric","type":"counter"}`,
			mockExpect: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), types.MetricID{ID: "missingMetric", Type: types.Counter}).
					Return(nil, nil)
			},
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockExpect()

			req := httptest.NewRequest(tt.method, tt.url, strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			defer req.Body.Close()

			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			if tt.expectedBody != "" {
				var got, want map[string]interface{}
				err1 := json.Unmarshal(rr.Body.Bytes(), &got)
				err2 := json.Unmarshal([]byte(tt.expectedBody), &want)
				assert.NoError(t, err1)
				assert.NoError(t, err2)
				assert.Equal(t, want, got)
			}
		})
	}
}
