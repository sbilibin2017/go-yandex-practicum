package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricListHTMLHandler_serveHTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	float64Ptr := func(f float64) *float64 { return &f }
	int64Ptr := func(i int64) *int64 { return &i }

	tests := []struct {
		name           string
		mockSetup      func(m *MockMetricLister)
		expectedCode   int
		expectedBody   string
		expectedNoBody bool
	}{
		{
			name: "success with metrics",
			mockSetup: func(m *MockMetricLister) {
				m.EXPECT().List(gomock.Any()).Return([]*types.Metrics{
					{
						ID:    "metric_gauge",
						MType: types.Gauge,
						Value: float64Ptr(3.14),
					},
					{
						ID:    "metric_counter",
						MType: types.Counter,
						Delta: int64Ptr(42),
					},
				}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: "<li>metric_gauge: 3.14</li>",
		},
		{
			name: "error listing metrics",
			mockSetup: func(m *MockMetricLister) {
				m.EXPECT().List(gomock.Any()).Return(nil, errors.New("db failure"))
			},
			expectedCode:   http.StatusInternalServerError,
			expectedNoBody: true,
		},
		{
			name: "success with empty metrics list",
			mockSetup: func(m *MockMetricLister) {
				m.EXPECT().List(gomock.Any()).Return([]*types.Metrics{}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: "<ul>",
		},
		{
			name: "metrics with nil values",
			mockSetup: func(m *MockMetricLister) {
				m.EXPECT().List(gomock.Any()).Return([]*types.Metrics{
					{ID: "gauge_nil", MType: types.Gauge, Value: nil},
					{ID: "counter_nil", MType: types.Counter, Delta: nil},
				}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: "<ul>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLister := NewMockMetricLister(ctrl)
			tt.mockSetup(mockLister)

			handler := NewMetricListHTMLHandler(
				WithMetricLister(mockLister),
			)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if req.Body != nil {
				defer req.Body.Close()
			}

			w := httptest.NewRecorder()

			handler.serveHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close() // <--- Close response body

			body := w.Body.String()

			assert.Equal(t, tt.expectedCode, resp.StatusCode)
			if tt.expectedNoBody {
				assert.Empty(t, body)
			} else {
				assert.Contains(t, body, tt.expectedBody)
			}
		})
	}
}

func TestMetricListHTMLHandler_RegisterRoute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := NewMockMetricLister(ctrl)
	mockLister.EXPECT().List(gomock.Any()).Return(nil, nil).AnyTimes()

	handler := NewMetricListHTMLHandler(WithMetricLister(mockLister))

	r := chi.NewRouter()
	handler.RegisterRoute(r)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if req.Body != nil {
		defer req.Body.Close()
	}

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close() // <--- Close response body

	// Since List returns nil,nil, the handler returns empty list page with 200 OK
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, w.Body.String(), "<ul>")
}
