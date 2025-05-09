package main

import (
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type MetricAPISuite struct {
	suite.Suite
	cmd    *exec.Cmd
	client *resty.Client
}

func (s *MetricAPISuite) SetupSuite() {
	s.cmd = exec.Command("../bin/server", "-a", ":8080", "-l", "debug")
	err := s.cmd.Start()
	require.NoError(s.T(), err)

	time.Sleep(500 * time.Millisecond)

	s.client = resty.New().
		SetBaseURL("http://localhost:8080").
		SetRedirectPolicy(resty.NoRedirectPolicy())
}

func (s *MetricAPISuite) TearDownSuite() {
	if s.cmd != nil {
		_ = s.cmd.Process.Kill()
	}
}

func (s *MetricAPISuite) TestMetricAPI() {
	testCases := []struct {
		name       string
		url        string
		statusCode int
	}{
		{
			name:       "valid gauge metric",
			url:        "/update/gauge/heap_alloc/123.45",
			statusCode: http.StatusOK,
		},
		{
			name:       "valid counter metric",
			url:        "/update/counter/requests/10",
			statusCode: http.StatusOK,
		},
		{
			name:       "missing metric name",
			url:        "/update/gauge//123.45",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "invalid metric type",
			url:        "/update/invalid_type/metric/123",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "invalid counter value",
			url:        "/update/counter/metric/not_a_number",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "invalid gauge value",
			url:        "/update/gauge/metric/not_a_float",
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			resp, err := s.client.R().
				SetHeader("Content-Type", "text/plain").
				Post(tc.url)

			require.NoError(t, err)
			assert.Equal(t, tc.statusCode, resp.StatusCode())
		})
	}
}

func TestMetricAPISuite(t *testing.T) {
	suite.Run(t, new(MetricAPISuite))
}
