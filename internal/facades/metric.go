package facades

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricFacade struct {
	client resty.Client
}

func NewMetricFacade(
	client resty.Client,
	flagServerAddress string,
) *MetricFacade {
	if !strings.HasPrefix(flagServerAddress, "http://") && !strings.HasPrefix(flagServerAddress, "https://") {
		flagServerAddress = "http://" + flagServerAddress
	}
	client.SetBaseURL(flagServerAddress)
	return &MetricFacade{client: client}
}

func (mf *MetricFacade) Update(
	ctx context.Context, req types.MetricUpdatePathRequest,
) error {
	url := fmt.Sprintf("/update/%s/%s/%s", req.Type, req.Name, req.Value)
	resp, err := mf.client.R().
		SetHeader("Content-Type", "text/plain").
		Post(url)
	if err != nil {
		return fmt.Errorf("failed to send metric %s: %v", req.Name, err)
	}
	if resp.IsError() {
		return fmt.Errorf("error response from server for metric %s: %s", req.Name, resp.String())
	}
	return nil
}
