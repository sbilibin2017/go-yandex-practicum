package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestMetricListAllHTMLHandler_ServeHTTP(t *testing.T) {
	floatPtr := func(f float64) *float64 { return &f }
	intPtr := func(i int64) *int64 { return &i }

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name            string
		mockReturn      []*types.Metrics
		mockError       error
		wantStatusCode  int
		wantBodySubstr  string
		wantContentType string
	}{
		{
			name: "success returns metrics html",
			mockReturn: []*types.Metrics{
				{ID: "metric1", MType: types.Gauge, Value: floatPtr(3.14)},
				{ID: "metric2", MType: types.Counter, Delta: intPtr(10)},
			},
			mockError:       nil,
			wantStatusCode:  http.StatusOK,
			wantBodySubstr:  "metric1",
			wantContentType: "text/html; charset=utf-8",
		},
		{
			name:           "service returns error",
			mockReturn:     nil,
			mockError:      assert.AnError,
			wantStatusCode: http.StatusInternalServerError,
			wantBodySubstr: "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := NewMockMetricListAllHTMLService(ctrl)
			mockSvc.EXPECT().ListAll(gomock.Any()).Return(tt.mockReturn, tt.mockError)

			handler := NewMetricListAllHTMLHandler(mockSvc)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			resp := w.Result()
			body := w.Body.String()

			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)

			if tt.wantContentType != "" {
				assert.Equal(t, tt.wantContentType, resp.Header.Get("Content-Type"))
			}

			if tt.wantBodySubstr != "" {
				assert.Contains(t, body, tt.wantBodySubstr)
			}
		})
	}
}
