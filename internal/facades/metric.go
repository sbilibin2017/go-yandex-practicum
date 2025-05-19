package facades

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricFacade struct {
	client        *resty.Client
	marshalerFunc func(v any) ([]byte, error)
	key           string
	header        string
	hashFunc      func(data []byte, key string) string
}

func NewMetricFacade(
	client *resty.Client,
	marshalerFunc func(v any) ([]byte, error),
	hashFunc func(data []byte, key string) string,
	serverAddress string,
	key string,
	header string,
) *MetricFacade {
	if !strings.HasPrefix(serverAddress, "http://") && !strings.HasPrefix(serverAddress, "https://") {
		serverAddress = "http://" + serverAddress
	}
	client = client.SetBaseURL(serverAddress)
	return &MetricFacade{
		client:        client,
		marshalerFunc: marshalerFunc,
		key:           key,
		header:        header,
		hashFunc:      hashFunc,
	}
}

func (mf *MetricFacade) Updates(
	ctx context.Context,
	metrics []types.Metrics,
) error {
	if len(metrics) == 0 {
		return nil
	}

	bodyBytes, err := mf.marshalerFunc(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	req := mf.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(bodyBytes)

	mf.setHashHeader(req, bodyBytes)

	resp, err := req.Post("/updates/")
	if err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("error response from server for metrics: %s", resp.String())
	}
	return nil
}

func (mf *MetricFacade) setHashHeader(req *resty.Request, body []byte) {
	if mf.key == "" || mf.hashFunc == nil {
		return
	}
	hashValue := mf.hashFunc(body, mf.key)
	req.SetHeader(mf.header, hashValue)
}
