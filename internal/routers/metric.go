package routers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewMetricRouter(
	metricUpdateHandler http.HandlerFunc,
	metricGetHandler http.HandlerFunc,
	metricListAllHandler http.HandlerFunc,
	loggingMiddleware func(next http.Handler) http.Handler,
) *chi.Mux {
	router := chi.NewRouter()

	router.Use(
		loggingMiddleware,
	)

	router.Post("/update/{type}/{name}/{value}", metricUpdateHandler)
	router.Get("/value/{type}/{name}", metricGetHandler)
	router.Get("/", metricListAllHandler)

	return router
}
