package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/sbilibin2017/go-yandex-practicum/internal/handlers"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestMetricListAllHTMLHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	float64Ptr := func(f float64) *float64 {
		return &f
	}

	int64Ptr := func(i int64) *int64 {
		return &i
	}

	mockSvc := handlers.NewMockMetricListAllHTMLService(ctrl)
	handler := handlers.NewMetricListAllHTMLHandler(mockSvc)

	tests := []struct {
		name          string
		mockReturn    []types.Metrics
		mockErr       error
		wantStatus    int
		wantBodyParts []string // substrings that must appear in response body
	}{
		{
			name: "Success with gauge and counter",
			mockReturn: []types.Metrics{
				{
					ID:    "gauge1",
					Type:  types.Gauge,
					Value: float64Ptr(3.14),
				},
				{
					ID:    "counter1",
					Type:  types.Counter,
					Delta: int64Ptr(42),
				},
			},
			mockErr:    nil,
			wantStatus: http.StatusOK,
			wantBodyParts: []string{
				"gauge1: 3.14",
				"counter1: 42",
				"<!DOCTYPE html>",
				"<ul>",
				"</ul>",
			},
		},
		{
			name:       "Service error returns 500",
			mockReturn: nil,
			mockErr:    assert.AnError,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc.EXPECT().ListAll(gomock.Any()).Return(tt.mockReturn, tt.mockErr)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			handler(w, req)

			resp := w.Result()
			defer resp.Body.Close() // <-- close the response body here

			bodyBytes := w.Body.Bytes()
			body := string(bodyBytes)

			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			for _, part := range tt.wantBodyParts {
				assert.Contains(t, body, part)
			}

			// When error returned, body contains default error message from http.Error
			if tt.mockErr != nil {
				assert.Contains(t, body, "Internal server error")
			}
		})
	}
}
