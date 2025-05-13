package apps

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/go-yandex-practicum/internal/handlers"
	"github.com/sbilibin2017/go-yandex-practicum/internal/middlewares"
	"github.com/sbilibin2017/go-yandex-practicum/internal/repositories"
	"github.com/sbilibin2017/go-yandex-practicum/internal/routers"
	"github.com/sbilibin2017/go-yandex-practicum/internal/services"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func ConfigureServerApp(
	ctx context.Context,
	metricsMap map[types.MetricID]types.Metrics,
	router *chi.Mux,
	server *http.Server,
) error {
	metricMemorySaveRepository := repositories.NewMetricMemorySaveRepository(metricsMap)
	metricMemoryGetByIDRepository := repositories.NewMetricMemoryGetByIDRepository(metricsMap)
	metricMemoryListAllRepository := repositories.NewMetricMemoryListAllRepository(metricsMap)

	metricUpdateService := services.NewMetricUpdateService(
		metricMemoryGetByIDRepository,
		metricMemorySaveRepository,
	)
	metricGetService := services.NewMetricGetService(
		metricMemoryGetByIDRepository,
	)
	metricListAllService := services.NewMetricListAllService(
		metricMemoryListAllRepository,
	)

	metricUpdatePathHandler := handlers.NewMetricUpdatePathHandler(metricUpdateService)
	metricUpdateBodyHandler := handlers.NewMetricUpdateBodyHandler(metricUpdateService)
	metricGetPathHandler := handlers.NewMetricGetPathHandler(metricGetService)
	metricGetBodyHandler := handlers.NewMetricGetBodyHandler(metricGetService)
	metricListAllHandler := handlers.NewMetricListAllHTMLHandler(metricListAllService)

	metricRouter := routers.NewMetricRouter(
		metricUpdatePathHandler,
		metricUpdateBodyHandler,
		metricGetPathHandler,
		metricGetBodyHandler,
		metricListAllHandler,
		middlewares.LoggingMiddleware,
	)

	router.Mount("/", metricRouter)

	server.Handler = router

	return nil
}
