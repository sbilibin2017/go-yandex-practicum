package facades

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricFacade struct {
	client        *resty.Client
	serverAddress string
	header        string
	key           string
}

func NewMetricFacade(
	client *resty.Client,
	serverAddress string,
	header string,
	key string,
) *MetricFacade {
	if !strings.HasPrefix(serverAddress, "http://") && !strings.HasPrefix(serverAddress, "https://") {
		serverAddress = "http://" + serverAddress
	}

	client.
		SetBaseURL(serverAddress).
		SetRetryCount(3).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return err != nil
		}).
		SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
			switch resp.Request.Attempt {
			case 1:
				return 1 * time.Second, nil
			case 2:
				return 3 * time.Second, nil
			case 3:
				return 5 * time.Second, nil
			default:
				return 0, nil
			}
		})

	return &MetricFacade{
		client:        client,
		serverAddress: serverAddress,
		header:        header,
		key:           key,
	}
}

func (mf *MetricFacade) Updates(
	ctx context.Context,
	metrics []types.Metrics,
) error {
	if len(metrics) == 0 {
		return nil
	}

	bodyBytes, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	req := mf.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(bodyBytes)

	if mf.key != "" {
		h := hmac.New(sha256.New, []byte(mf.key))
		h.Write(bodyBytes)
		hashSum := hex.EncodeToString(h.Sum(nil))
		req.SetHeader(mf.header, hashSum)
	}

	resp, err := req.Post("/updates/")
	if err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("error response from server for metrics: %s", resp.String())
	}
	return nil
}
