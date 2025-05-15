package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum/internal/handlers"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestMetricUpdatePathHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := handlers.NewMockMetricUpdatePathService(ctrl)
	handler := handlers.NewMetricUpdatePathHandler(mockService)

	router := chi.NewRouter()
	router.Post("/update/{type}/{name}/{value}", handler)

	tests := []struct {
		name           string
		url            string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name: "Valid Counter Metric",
			url:  "/update/counter/counter1/10",
			mockSetup: func() {
				mockService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Valid Gauge Metric",
			url:  "/update/gauge/gauge1/10.5",
			mockSetup: func() {
				mockService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing Metric Name",
			url:            "/update/counter//10",
			mockSetup:      func() {},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid Metric Type",
			url:            "/update/invalid/invalid1/10",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid Value for Counter",
			url:            "/update/counter/counter1/invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid Value for Gauge",
			url:            "/update/gauge/gauge1/invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Update Error",
			url:  "/update/counter/counter1/10",
			mockSetup: func() {
				mockService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(types.ErrMetricInternal).Times(1)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodPost, tt.url, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			assert.Equal(t, tt.expectedStatus, rec.Code)

		})
	}
}
