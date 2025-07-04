package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	pb "github.com/sbilibin2017/go-yandex-practicum/protos"
	"github.com/stretchr/testify/assert"
)

func TestMetricUpdatePathHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUpdater := NewMockMetricUpdater(ctrl)
	handler := NewMetricUpdatePathHandler(WithMetricUpdaterPath(mockUpdater))

	r := chi.NewRouter()
	handler.RegisterRoute(r)

	tests := []struct {
		name         string
		method       string
		url          string
		mockExpect   func()
		expectedCode int
	}{
		{
			name:   "Valid counter metric",
			method: http.MethodPost,
			url:    "/update/counter/myCounter/100",
			mockExpect: func() {
				mockUpdater.EXPECT().
					Updates(gomock.Any(), gomock.AssignableToTypeOf([]*types.Metrics{})).
					Return(nil, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:   "Valid gauge metric",
			method: http.MethodPost,
			url:    "/update/gauge/myGauge/99.9",
			mockExpect: func() {
				mockUpdater.EXPECT().
					Updates(gomock.Any(), gomock.AssignableToTypeOf([]*types.Metrics{})).
					Return(nil, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "Missing metric name",
			method:       http.MethodPost,
			url:          "/update/counter//100", // empty name param triggers 404
			mockExpect:   func() {},
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "Missing metric value",
			method:       http.MethodPost,
			url:          "/update/gauge/myGauge", // no value param triggers 400
			mockExpect:   func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Invalid counter value",
			method:       http.MethodPost,
			url:          "/update/counter/myCounter/notanumber",
			mockExpect:   func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Invalid gauge value",
			method:       http.MethodPost,
			url:          "/update/gauge/myGauge/invalidfloat",
			mockExpect:   func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Unsupported metric type",
			method:       http.MethodPost,
			url:          "/update/unknown/myMetric/123",
			mockExpect:   func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:   "Updater returns error",
			method: http.MethodPost,
			url:    "/update/counter/myMetric/123",
			mockExpect: func() {
				mockUpdater.EXPECT().
					Updates(gomock.Any(), gomock.AssignableToTypeOf([]*types.Metrics{})).
					Return(nil, context.DeadlineExceeded)
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockExpect()

			req := httptest.NewRequest(tt.method, tt.url, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
		})
	}
}

func TestMetricUpdateBodyHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUpdater := NewMockMetricUpdater(ctrl)
	handler := NewMetricUpdateBodyHandler(WithMetricUpdaterBody(mockUpdater))

	r := chi.NewRouter()
	handler.RegisterRoute(r)

	ptrInt64 := func(i int64) *int64 { return &i }
	ptrFloat64 := func(f float64) *float64 { return &f }

	tests := []struct {
		name         string
		payload      any
		mockExpect   func()
		expectedCode int
		expectedBody *types.Metrics
	}{
		{
			name: "Valid counter metric",
			payload: types.Metrics{
				ID:    "myCounter",
				Type:  types.Counter,
				Delta: ptrInt64(100),
			},
			mockExpect: func() {
				mockUpdater.EXPECT().
					Updates(gomock.Any(), gomock.AssignableToTypeOf([]*types.Metrics{})).
					Return([]*types.Metrics{
						{
							ID:    "myCounter",
							Type:  types.Counter,
							Delta: ptrInt64(100),
						},
					}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: &types.Metrics{
				ID:    "myCounter",
				Type:  types.Counter,
				Delta: ptrInt64(100),
			},
		},
		{
			name: "Valid gauge metric",
			payload: types.Metrics{
				ID:    "myGauge",
				Type:  types.Gauge,
				Value: ptrFloat64(99.9),
			},
			mockExpect: func() {
				mockUpdater.EXPECT().
					Updates(gomock.Any(), gomock.AssignableToTypeOf([]*types.Metrics{})).
					Return([]*types.Metrics{
						{
							ID:    "myGauge",
							Type:  types.Gauge,
							Value: ptrFloat64(99.9),
						},
					}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: &types.Metrics{
				ID:    "myGauge",
				Type:  types.Gauge,
				Value: ptrFloat64(99.9),
			},
		},
		{
			name:         "Invalid JSON",
			payload:      `{"id": "metric1", "Type": "counter", "delta": "notanumber"}`,
			mockExpect:   func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Missing metric ID",
			payload:      types.Metrics{Type: types.Counter, Delta: ptrInt64(1)},
			mockExpect:   func() {},
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "Counter metric missing Delta",
			payload:      types.Metrics{ID: "myCounter", Type: types.Counter},
			mockExpect:   func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Gauge metric missing Value",
			payload:      types.Metrics{ID: "myGauge", Type: types.Gauge},
			mockExpect:   func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Unsupported metric type",
			payload: types.Metrics{
				ID:   "myMetric",
				Type: "unsupported",
			},
			mockExpect:   func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Updater returns error",
			payload: types.Metrics{
				ID:    "myCounter",
				Type:  types.Counter,
				Delta: ptrInt64(123),
			},
			mockExpect: func() {
				mockUpdater.EXPECT().
					Updates(gomock.Any(), gomock.AssignableToTypeOf([]*types.Metrics{})).
					Return(nil, context.DeadlineExceeded)
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockExpect()

			var bodyBytes []byte
			var err error
			switch v := tt.payload.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, err = json.Marshal(tt.payload)
				if err != nil {
					t.Fatalf("failed to marshal payload: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(bodyBytes))
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			if tt.expectedCode == http.StatusOK && tt.expectedBody != nil {
				var gotMetric types.Metrics
				err := json.NewDecoder(rr.Body).Decode(&gotMetric)
				assert.NoError(t, err)
				assert.Equal(t, *tt.expectedBody, gotMetric)
			}
		})
	}
}

func TestMetricUpdatesBodyHandler(t *testing.T) {
	ptrInt64 := func(i int64) *int64 { return &i }
	ptrFloat64 := func(f float64) *float64 { return &f }

	tests := []struct {
		name         string
		payload      any
		mockExpect   func(mockUpdater *MockMetricUpdater)
		expectedCode int
		expectedBody []*types.Metrics
	}{
		{
			name: "Valid batch update",
			payload: []*types.Metrics{
				{ID: "counter1", Type: types.Counter, Delta: ptrInt64(10)},
				{ID: "gauge1", Type: types.Gauge, Value: ptrFloat64(3.14)},
				{ID: "counter2", Type: types.Counter, Delta: ptrInt64(20)},
			},
			mockExpect: func(mockUpdater *MockMetricUpdater) {
				mockUpdater.EXPECT().
					Updates(gomock.Any(), gomock.AssignableToTypeOf([]*types.Metrics{})).
					Return([]*types.Metrics{
						{ID: "counter1", Type: types.Counter, Delta: ptrInt64(10)},
						{ID: "gauge1", Type: types.Gauge, Value: ptrFloat64(3.14)},
						{ID: "counter2", Type: types.Counter, Delta: ptrInt64(20)},
					}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: []*types.Metrics{
				{ID: "counter1", Type: types.Counter, Delta: ptrInt64(10)},
				{ID: "gauge1", Type: types.Gauge, Value: ptrFloat64(3.14)},
				{ID: "counter2", Type: types.Counter, Delta: ptrInt64(20)},
			},
		},
		{
			name:         "Invalid JSON",
			payload:      `[{"id":"counter1","Type":"counter","delta":"notanumber"}]`,
			mockExpect:   func(_ *MockMetricUpdater) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:    "Empty payload",
			payload: []*types.Metrics{},
			mockExpect: func(_ *MockMetricUpdater) {
				// No calls expected, handler returns 400 before calling Updates
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Missing metric ID",
			payload: []*types.Metrics{
				{Type: types.Counter, Delta: ptrInt64(1)},
			},
			mockExpect: func(_ *MockMetricUpdater) {
				// No calls expected, handler returns 404 before calling Updates
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name: "Counter metric missing Delta",
			payload: []*types.Metrics{
				{ID: "counter1", Type: types.Counter},
			},
			mockExpect: func(_ *MockMetricUpdater) {
				// No calls expected, handler returns 400 before calling Updates
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Gauge metric missing Value",
			payload: []*types.Metrics{
				{ID: "gauge1", Type: types.Gauge},
			},
			mockExpect: func(_ *MockMetricUpdater) {
				// No calls expected, handler returns 400 before calling Updates
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Unsupported metric type",
			payload: []*types.Metrics{
				{ID: "metricX", Type: "unsupported"},
			},
			mockExpect: func(_ *MockMetricUpdater) {
				// No calls expected, handler returns 400 before calling Updates
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Updater returns error",
			payload: []*types.Metrics{
				{ID: "counter1", Type: types.Counter, Delta: ptrInt64(123)},
			},
			mockExpect: func(mockUpdater *MockMetricUpdater) {
				mockUpdater.EXPECT().
					Updates(gomock.Any(), gomock.AssignableToTypeOf([]*types.Metrics{})).
					Return(nil, context.DeadlineExceeded)
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUpdater := NewMockMetricUpdater(ctrl)
			handler := NewMetricUpdatesBodyHandler(WithMetricUpdaterBatchBody(mockUpdater))

			r := chi.NewRouter()
			handler.RegisterRoute(r)

			tt.mockExpect(mockUpdater)

			var bodyBytes []byte
			var err error
			switch v := tt.payload.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, err = json.Marshal(tt.payload)
				if err != nil {
					t.Fatalf("failed to marshal payload: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(bodyBytes))
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			if tt.expectedCode == http.StatusOK && tt.expectedBody != nil {
				var gotMetrics []*types.Metrics
				err := json.NewDecoder(rr.Body).Decode(&gotMetrics)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, gotMetrics)
			}
		})
	}
}

func TestMetricUpdaterService_Updates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ptrFloat64 := func(f float64) *float64 { return &f }
	ptrInt64 := func(i int64) *int64 { return &i }
	ctx := context.Background()

	tests := []struct {
		name         string
		inputMetrics []*pb.Metric
		setup        func(mockUpdater *MockMetricUpdater)
		wantErr      string
		wantLen      int
	}{
		{
			name: "successful update with gauge and counter",
			inputMetrics: []*pb.Metric{
				{Id: "m1", Type: "gauge", Value: 1.23, Delta: 0},
				{Id: "m2", Type: "counter", Value: 0, Delta: 42},
			},
			setup: func(mockUpdater *MockMetricUpdater) {
				expectedMetrics := []*types.Metrics{
					{
						ID:    "m1",
						Type:  "gauge",
						Value: ptrFloat64(1.23),
						Delta: ptrInt64(0),
					},
					{
						ID:    "m2",
						Type:  "counter",
						Value: ptrFloat64(0),
						Delta: ptrInt64(42),
					},
				}
				mockUpdater.EXPECT().
					Updates(ctx, gomock.Eq(expectedMetrics)).
					Return(expectedMetrics, nil)
			},
			wantErr: "",
			wantLen: 2,
		},
		{
			name: "error returned from updater",
			inputMetrics: []*pb.Metric{
				{Id: "m3", Type: "gauge", Value: 5.67, Delta: 0},
			},
			setup: func(mockUpdater *MockMetricUpdater) {
				expectedMetrics := []*types.Metrics{
					{
						ID:    "m3",
						Type:  "gauge",
						Value: ptrFloat64(5.67),
						Delta: ptrInt64(0),
					},
				}
				mockUpdater.EXPECT().
					Updates(ctx, gomock.Eq(expectedMetrics)).
					Return(nil, errors.New("update failed"))
			},
			wantErr: "update failed",
			wantLen: 0,
		},
		{
			name:         "empty input metrics",
			inputMetrics: []*pb.Metric{},
			setup: func(mockUpdater *MockMetricUpdater) {
				mockUpdater.EXPECT().
					Updates(ctx, gomock.Eq([]*types.Metrics{})).
					Return([]*types.Metrics{}, nil)
			},
			wantErr: "",
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUpdater := NewMockMetricUpdater(ctrl)
			tt.setup(mockUpdater) // setup mocks per test case

			service := &MetricGRPCUpdaterHandler{
				updater: mockUpdater,
			}

			req := &pb.UpdateMetricsRequest{Metrics: tt.inputMetrics}
			resp, err := service.Updates(ctx, req)

			if tt.wantErr != "" {
				assert.NoError(t, err) // your code returns nil error but embeds error in response
				assert.NotNil(t, resp)
				assert.Equal(t, tt.wantErr, resp.Error)
				assert.Len(t, resp.Metrics, 0)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Empty(t, resp.Error)
			assert.Len(t, resp.Metrics, tt.wantLen)

			for i, m := range resp.Metrics {
				assert.Equal(t, tt.inputMetrics[i].Id, m.Id)
				assert.Equal(t, tt.inputMetrics[i].Type, m.Type)
				assert.Equal(t, tt.inputMetrics[i].Value, m.Value)
				assert.Equal(t, tt.inputMetrics[i].Delta, m.Delta)
			}
		})
	}
}
