package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/sbilibin2017/go-yandex-practicum/internal/handlers"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestMetricUpdateBodyHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := handlers.NewMockMetricUpdateBodyService(ctrl)
	handler := handlers.NewMetricUpdateBodyHandler(mockSvc)

	tests := []struct {
		name           string
		input          any
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:           "Invalid JSON",
			input:          "not-json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing ID",
			input: types.Metrics{
				Type:  types.Counter,
				Delta: func() *int64 { v := int64(1); return &v }(),
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Counter missing Delta",
			input: types.Metrics{
				ID:   "metric1",
				Type: types.Counter,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Gauge missing Value",
			input: types.Metrics{
				ID:   "metric1",
				Type: types.Gauge,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid Type",
			input: types.Metrics{
				ID:   "metric1",
				Type: "invalid",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service update error",
			input: types.Metrics{
				ID:   "metric1",
				Type: types.Counter,
				Delta: func() *int64 {
					v := int64(123)
					return &v
				}(),
			},
			mockSetup: func() {
				mockSvc.EXPECT().
					Updates(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Success Counter",
			input: types.Metrics{
				ID:   "metric1",
				Type: types.Counter,
				Delta: func() *int64 {
					v := int64(42)
					return &v
				}(),
			},
			mockSetup: func() {
				mockSvc.EXPECT().
					Updates(gomock.Any(), gomock.Any()).
					Return([]*types.Metrics{{
						ID:    "metric1",
						Type:  types.Counter,
						Delta: func() *int64 { v := int64(42); return &v }(),
					}}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Success Gauge",
			input: types.Metrics{
				ID:    "metric2",
				Type:  types.Gauge,
				Value: func() *float64 { v := 3.14; return &v }(),
			},
			mockSetup: func() {
				mockSvc.EXPECT().
					Updates(gomock.Any(), gomock.Any()).
					Return([]*types.Metrics{{
						ID:    "metric2",
						Type:  types.Gauge,
						Value: func() *float64 { v := 3.14; return &v }(),
					}}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			var bodyBytes []byte
			var err error
			switch v := tt.input.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, err = json.Marshal(v)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			handler(w, req)

			resp := w.Result()
			defer resp.Body.Close() // Close response body to avoid leaks

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

type errorWriter struct{}

func (e *errorWriter) Header() http.Header        { return http.Header{} }
func (e *errorWriter) Write([]byte) (int, error)  { return 0, errors.New("write error") }
func (e *errorWriter) WriteHeader(statusCode int) {}

func TestEncodeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := handlers.NewMockMetricUpdateBodyService(ctrl)
	handler := handlers.NewMetricUpdateBodyHandler(mockSvc)

	metric := types.Metrics{
		ID:   "metric1",
		Type: types.Counter,
		Delta: func() *int64 {
			v := int64(123)
			return &v
		}(),
	}

	mockSvc.EXPECT().
		Updates(gomock.Any(), gomock.Any()).
		Return([]*types.Metrics{&metric}, nil)

	bodyBytes, err := json.Marshal(metric)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, "/update", io.NopCloser(bytes.NewReader(bodyBytes)))
	if err != nil {
		t.Fatal(err)
	}
	defer req.Body.Close() // Close request body

	ew := &errorWriter{}

	handler(ew, req)
	// No panic means test passed
}
