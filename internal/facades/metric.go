package facades

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/go-yandex-practicum/internal/logger"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"go.uber.org/zap"
)

type MetricFacade struct {
	client   *resty.Client
	endpoint string
}

func NewMetricFacade(client *resty.Client, flagServerAddress string, endpoint string) *MetricFacade {
	if !strings.HasPrefix(flagServerAddress, "http://") && !strings.HasPrefix(flagServerAddress, "https://") {
		flagServerAddress = "http://" + flagServerAddress
	}
	client.SetBaseURL(flagServerAddress)
	return &MetricFacade{client: client, endpoint: endpoint}
}

func (mf *MetricFacade) Update(ctx context.Context, metrics types.Metrics) error {
	resp, err := mf.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(metrics).
		Post(mf.endpoint)
	if err != nil {
		logger.Log.Error("Failed to send metric",
			zap.String("metric_id", string(metrics.ID)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send metric %s: %v", metrics.ID, err)
	}
	if resp.IsError() {
		logger.Log.Error("Server returned error response",
			zap.String("metric_id", string(metrics.ID)),
			zap.Int("status_code", resp.StatusCode()),
			zap.String("response_body", resp.String()),
		)
		return fmt.Errorf("error response from server for metric %s: %s", metrics.ID, resp.String())
	}
	return nil
}
