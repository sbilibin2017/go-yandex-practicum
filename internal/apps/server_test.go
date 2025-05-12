package apps

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestConfigureServerApp(t *testing.T) {
	ctx := context.Background()
	metricsMap := make(map[types.MetricID]types.Metrics)
	router := chi.NewRouter()
	server := &http.Server{}

	err := ConfigureServerApp(ctx, metricsMap, router, server)
	assert.NoError(t, err)

}
