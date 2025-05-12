package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestMetricUpdatePathHandler_WithRouter(t *testing.T) {
	tests := []struct {
		name           string
		mtype          string
		nameParam      string
		valueParam     string
		mockUpdateErr  error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Missing metric value (empty string)",
			mtype:          string(types.CounterMetricType),
			nameParam:      "metric1",
			valueParam:     "",
			mockUpdateErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Metric value is required\n",
		},
		{
			name:           "Valid counter metric update",
			mtype:          string(types.CounterMetricType),
			nameParam:      "metric1",
			valueParam:     "10",
			mockUpdateErr:  nil,
			expectedStatus: http.StatusOK,
			expectedBody:   "Metric updated successfully",
		},
		{
			name:           "Valid gauge metric update",
			mtype:          string(types.GaugeMetricType),
			nameParam:      "metric2",
			valueParam:     "3.14",
			mockUpdateErr:  nil,
			expectedStatus: http.StatusOK,
			expectedBody:   "Metric updated successfully",
		},
		{
			name:           "Invalid metric type",
			mtype:          "invalid",
			nameParam:      "metric1",
			valueParam:     "10",
			mockUpdateErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid metric type\n",
		},
		{
			name:           "Missing metric name",
			mtype:          string(types.CounterMetricType),
			nameParam:      "",
			valueParam:     "10",
			mockUpdateErr:  nil,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Metric name is required\n",
		},
		{
			name:           "Missing metric value",
			mtype:          string(types.CounterMetricType),
			nameParam:      "metric1",
			valueParam:     "",
			mockUpdateErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Metric value is required\n",
		},
		{
			name:           "Invalid value for counter metric",
			mtype:          string(types.CounterMetricType),
			nameParam:      "metric1",
			valueParam:     "invalid",
			mockUpdateErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid metric value for counter\n",
		},
		{
			name:           "Invalid value for gauge metric",
			mtype:          string(types.GaugeMetricType),
			nameParam:      "metric2",
			valueParam:     "invalid",
			mockUpdateErr:  nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid metric value for gauge\n",
		},
		{
			name:           "Update service error",
			mtype:          string(types.CounterMetricType),
			nameParam:      "metric3",
			valueParam:     "42",
			mockUpdateErr:  errors.New("update failed"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Metric not updated\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := NewMockMetricUpdatePathService(ctrl)

			if tt.expectedStatus == http.StatusOK || tt.expectedStatus == http.StatusInternalServerError {
				mockService.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(tt.mockUpdateErr).
					Times(1)
			}

			handler := NewMetricUpdatePathHandler(mockService)

			router := chi.NewRouter()
			router.Post("/update/{type}/{name}/{value}", handler)
			router.Post("/update/{type}/{name}", handler)

			var target string
			if tt.valueParam != "" {
				target = "/update/" + tt.mtype + "/" + tt.nameParam + "/" + tt.valueParam
			} else {
				target = "/update/" + tt.mtype + "/" + tt.nameParam
			}

			req, err := http.NewRequest("POST", target, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())
		})
	}
}
