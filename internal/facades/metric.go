package facades

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

type MetricFacade struct {
	client resty.Client
}

func NewMetricFacade(
	client resty.Client,
	flagServerURL string,
) *MetricFacade {
	return &MetricFacade{client: *client.SetBaseURL(flagServerURL)}
}

func (mf *MetricFacade) Update(
	ctx context.Context, metric map[string]any,
) error {
	url := fmt.Sprintf("/update/%s/%s/%s", metric["type"], metric["name"], fmt.Sprint(metric["value"]))
	resp, err := mf.client.R().
		SetHeader("Content-Type", "text/plain").
		Post(url)
	if err != nil {
		return fmt.Errorf("failed to send metric %s: %v", metric["name"], err)
	}
	if resp.IsError() {
		return fmt.Errorf("error response from server for metric %s: %s", metric["name"], resp.String())
	}
	return nil
}
