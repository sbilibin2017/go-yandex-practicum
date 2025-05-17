package facades

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricFacade struct {
	client *resty.Client
	key    string
	header string
}

func NewMetricFacade(client *resty.Client, flagServerAddress string, key string, header string) *MetricFacade {
	if !strings.HasPrefix(flagServerAddress, "http://") && !strings.HasPrefix(flagServerAddress, "https://") {
		flagServerAddress = "http://" + flagServerAddress
	}
	client = client.SetBaseURL(flagServerAddress)
	return &MetricFacade{
		client: client,
		key:    key,
		header: header,
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

	setHashHeader(req, bodyBytes, mf.key, mf.header)

	resp, err := req.Post("/updates/")
	if err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("error response from server for metrics: %s", resp.String())
	}
	return nil
}

func setHashHeader(req *resty.Request, body []byte, key string, header string) {
	if key == "" {
		return
	}
	h := hmac.New(sha256.New, []byte(key))
	h.Write(body)
	hashValue := hex.EncodeToString(h.Sum(nil))
	req.SetHeader(header, hashValue)
}
