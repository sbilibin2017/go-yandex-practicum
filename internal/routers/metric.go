package routers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewMetricRouter(
	metricUpdatePathHandler http.HandlerFunc,
	metricUpdateBodyHandler http.HandlerFunc,
	metricGetPathHandler http.HandlerFunc,
	metricGetBodyHandler http.HandlerFunc,
	metricListAllHandler http.HandlerFunc,
	loggingMiddleware func(next http.Handler) http.Handler,
	gzipMiddleware func(next http.Handler) http.Handler,
) *chi.Mux {
	router := chi.NewRouter()

	router.Use(
		loggingMiddleware,
		gzipMiddleware,
	)

	router.Post("/update/{type}/{name}/{value}", metricUpdatePathHandler)
	router.Post("/update/", metricUpdateBodyHandler)
	router.Get("/value/{type}/{name}", metricGetPathHandler)
	router.Post("/value/", metricGetBodyHandler)
	router.Get("/", metricListAllHandler)

	return router
}
