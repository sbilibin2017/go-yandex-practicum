package apps

import (
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func TestConfigureAgentApp(t *testing.T) {
	client := *resty.New()
	flagServerAddress := "localhost:8080"
	metricCh := make(chan types.MetricUpdatePathRequest, 1)
	pollTicker := time.NewTicker(time.Second)
	reportTicker := time.NewTicker(2 * time.Second)

	defer pollTicker.Stop()
	defer reportTicker.Stop()

	agent := ConfigureAgentApp(client, flagServerAddress, metricCh, pollTicker, reportTicker)

	assert.NotNil(t, agent)

}
