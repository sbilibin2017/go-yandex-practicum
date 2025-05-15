package facades

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricFacade struct {
	client *resty.Client
}

func NewMetricFacade(client *resty.Client, flagServerAddress string) *MetricFacade {
	if !strings.HasPrefix(flagServerAddress, "http://") && !strings.HasPrefix(flagServerAddress, "https://") {
		flagServerAddress = "http://" + flagServerAddress
	}
	client = client.SetBaseURL(flagServerAddress)
	return &MetricFacade{
		client: client,
	}
}

func (mf *MetricFacade) Updates(ctx context.Context, metrics []types.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}
	resp, err := mf.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(metrics).
		Post("/updates/")
	if err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("error response from server for metrics: %s", resp.String())
	}
	return nil
}
