package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestMetricUpdatesBodyHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockMetricUpdatesBodyService(ctrl)
	handler := NewMetricUpdatesBodyHandler(mockSvc)

	counterVal := int64(42)
	gaugeVal := float64(3.14)

	tests := []struct {
		name           string
		input          interface{}
		mockSetup      func()
		wantStatusCode int
	}{
		{
			name:           "invalid json",
			input:          "invalid json",
			mockSetup:      func() {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "missing id",
			input: []*types.Metrics{
				{ID: "", Type: types.Counter, Delta: &counterVal},
			},
			mockSetup:      func() {},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "counter missing delta",
			input: []*types.Metrics{
				{ID: "metric1", Type: types.Counter, Delta: nil},
			},
			mockSetup:      func() {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "gauge missing value",
			input: []*types.Metrics{
				{ID: "metric2", Type: types.Gauge, Value: nil},
			},
			mockSetup:      func() {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "invalid metric type",
			input: []*types.Metrics{
				{ID: "metric3", Type: "invalid"},
			},
			mockSetup:      func() {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "service update error",
			input: []*types.Metrics{
				{ID: "metric4", Type: types.Counter, Delta: &counterVal},
			},
			mockSetup: func() {
				mockSvc.EXPECT().
					Updates(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("update failed"))
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "success",
			input: []*types.Metrics{
				{ID: "metric5", Type: types.Counter, Delta: &counterVal},
				{ID: "metric6", Type: types.Gauge, Value: &gaugeVal},
			},
			mockSetup: func() {
				mockSvc.EXPECT().
					Updates(gomock.Any(), gomock.Any()).
					Return([]*types.Metrics{
						{ID: "metric5", Type: types.Counter, Delta: &counterVal},
						{ID: "metric6", Type: types.Gauge, Value: &gaugeVal},
					}, nil)
			},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			var body []byte
			var err error

			switch v := tt.input.(type) {
			case string:
				body = []byte(v)
			default:
				body, err = json.Marshal(v)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/metrics/update", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler(rec, req)

			resp := rec.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
		})
	}
}

type failingWriter struct{}

func (f *failingWriter) Header() http.Header        { return http.Header{} }
func (f *failingWriter) Write([]byte) (int, error)  { return 0, errFailWrite }
func (f *failingWriter) WriteHeader(statusCode int) {}

var errFailWrite = &writeError{}

type writeError struct{}

func (e *writeError) Error() string { return "write error" }

func TestEncodeErrorPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockMetricUpdatesBodyService(ctrl)
	handler := NewMetricUpdatesBodyHandler(mockSvc)

	delta := int64(42)
	metrics := []*types.Metrics{
		{ID: "id1", Type: types.Counter, Delta: &delta},
	}

	mockSvc.EXPECT().
		Updates(gomock.Any(), gomock.Any()).
		Return(metrics, nil)

	body, _ := json.Marshal(metrics)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))

	fw := &failingWriter{}
	handler(fw, req)

	// If no panic or deadlock, test passed.
}
