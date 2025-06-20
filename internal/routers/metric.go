package routers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewMetricRouter(
	metricUpdatePathHandler http.Handler,
	metricUpdateBodyHandler http.Handler,
	metricUpdatesBodyHandler http.Handler,
	metricGetPathHandler http.Handler,
	metricGetBodyHandler http.Handler,
	metricListAllHTMLHandler http.Handler,
	middlewares ...func(http.Handler) http.Handler,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middlewares...)

	r.Post("/update/{type}/{name}/{value}", metricUpdatePathHandler.ServeHTTP)
	r.Post("/update/{type}/{name}", metricUpdatePathHandler.ServeHTTP)
	r.Post("/update/", metricUpdateBodyHandler.ServeHTTP)
	r.Post("/updates/", metricUpdatesBodyHandler.ServeHTTP)
	r.Get("/value/{type}/{name}", metricGetPathHandler.ServeHTTP)
	r.Post("/value/", metricGetBodyHandler.ServeHTTP)
	r.Get("/", metricListAllHTMLHandler.ServeHTTP)

	return r
}
