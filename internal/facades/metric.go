package facades

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/go-yandex-practicum/internal/hash"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricFacade struct {
	client *resty.Client
	key    string
}

func NewMetricFacade(client *resty.Client, flagServerAddress, key string) *MetricFacade {
	if !strings.HasPrefix(flagServerAddress, "http://") && !strings.HasPrefix(flagServerAddress, "https://") {
		flagServerAddress = "http://" + flagServerAddress
	}
	client = client.SetBaseURL(flagServerAddress)
	return &MetricFacade{
		client: client,
		key:    key,
	}
}

func (mf *MetricFacade) Updates(ctx context.Context, metrics []types.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	bodyBytes, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	req := mf.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(bodyBytes)

	setHashHeader(req, bodyBytes, mf.key)

	resp, err := req.Post("/updates/")
	if err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("error response from server for metrics: %s", resp.String())
	}
	return nil
}

func setHashHeader(req *resty.Request, body []byte, key string) {
	if key == "" {
		return
	}
	hashValue := hash.HashWithKey(body, key)
	req.SetHeader(hash.Header, hashValue)
}
