package apps

import (
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/go-yandex-practicum/internal/facades"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/sbilibin2017/go-yandex-practicum/internal/workers"
)

func ConfigureAgentApp(
	client resty.Client,
	flagServerAddress string,
	metricCh chan types.MetricUpdatePathRequest,
	pollTicker *time.Ticker,
	reportTicker *time.Ticker,
) *workers.MetricAgent {
	metricFacade := facades.NewMetricFacade(client, flagServerAddress)

	metricAgent := workers.NewMetricAgent(
		metricFacade,
		metricCh,
		pollTicker,
		reportTicker,
	)

	return metricAgent
}
