package facades

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
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
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(metrics).
		Post(mf.endpoint)

	if err != nil {
		return fmt.Errorf("failed to send metric %s: %v", metrics.ID, err)
	}
	if resp.IsError() {
		return fmt.Errorf("error response from server for metric %s: %s", metrics.ID, resp.String())
	}
	return nil
}
