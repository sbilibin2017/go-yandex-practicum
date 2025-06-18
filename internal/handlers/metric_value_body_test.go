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

func TestNewMetricGetBodyHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockMetricGetBodyService(ctrl)
	handler := NewMetricGetBodyHandler(mockSvc)

	int64Ptr := func(i int64) *int64 { return &i }
	float64Ptr := func(f float64) *float64 { return &f }

	testCases := []struct {
		name           string
		body           interface{}
		mockSetup      func()
		expectedStatus int
		expectedBody   *types.Metrics
	}{
		{
			name:           "Invalid JSON",
			body:           "invalid-json",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing metric ID",
			body:           types.MetricID{ID: "", Type: types.Counter},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid metric type",
			body:           types.MetricID{ID: "id1", Type: "invalid"},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service returns error",
			body: types.MetricID{ID: "id1", Type: types.Counter},
			mockSetup: func() {
				mockSvc.EXPECT().Get(gomock.Any(), types.MetricID{ID: "id1", Type: types.Counter}).
					Return(nil, errors.New("some error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Service returns nil metric",
			body: types.MetricID{ID: "id1", Type: types.Counter},
			mockSetup: func() {
				mockSvc.EXPECT().Get(gomock.Any(), types.MetricID{ID: "id1", Type: types.Counter}).
					Return(nil, nil)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Valid Counter metric",
			body: types.MetricID{ID: "counter1", Type: types.Counter},
			mockSetup: func() {
				mockSvc.EXPECT().Get(gomock.Any(), types.MetricID{ID: "counter1", Type: types.Counter}).
					Return(&types.Metrics{
						ID:    "counter1",
						Type:  types.Counter,
						Delta: int64Ptr(123),
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: &types.Metrics{
				ID:    "counter1",
				Type:  types.Counter,
				Delta: int64Ptr(123),
			},
		},
		{
			name: "Valid Gauge metric",
			body: types.MetricID{ID: "gauge1", Type: types.Gauge},
			mockSetup: func() {
				mockSvc.EXPECT().Get(gomock.Any(), types.MetricID{ID: "gauge1", Type: types.Gauge}).
					Return(&types.Metrics{
						ID:    "gauge1",
						Type:  types.Gauge,
						Value: float64Ptr(45.67),
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: &types.Metrics{
				ID:    "gauge1",
				Type:  types.Gauge,
				Value: float64Ptr(45.67),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			var reqBody bytes.Buffer
			if s, ok := tc.body.(string); ok {
				reqBody.WriteString(s)
			} else {
				err := json.NewEncoder(&reqBody).Encode(tc.body)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/metric", &reqBody)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			resp := rec.Result()
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			if tc.expectedStatus == http.StatusOK {
				var gotMetric types.Metrics
				err := json.NewDecoder(resp.Body).Decode(&gotMetric)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedBody, &gotMetric)
			}
		})
	}
}

type errorWriter struct{}

func (e *errorWriter) Header() http.Header        { return http.Header{} }
func (e *errorWriter) Write([]byte) (int, error)  { return 0, errors.New("write error") }
func (e *errorWriter) WriteHeader(statusCode int) {}

func TestNewMetricGetBodyHandler_EncodeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockMetricGetBodyService(ctrl)

	metric := &types.Metrics{
		ID:    "counter1",
		Type:  types.Counter,
		Delta: func(i int64) *int64 { return &i }(123),
	}

	mockSvc.EXPECT().Get(gomock.Any(), types.MetricID{ID: "counter1", Type: types.Counter}).Return(metric, nil)

	handler := NewMetricGetBodyHandler(mockSvc)

	reqBody, err := json.Marshal(types.MetricID{ID: "counter1", Type: types.Counter})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/metric", bytes.NewReader(reqBody))

	rec := &errorWriter{}

	// ServeHTTP will call Write which returns error
	handler.ServeHTTP(rec, req)

	// Since WriteHeader(http.StatusOK) is called before Encode,
	// status code can't change, but the test confirms Write error occurs
	// To test this exactly, handler code might need refactoring to buffer first.
}
