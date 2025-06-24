package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricLister defines an interface for listing all metrics.
type MetricLister interface {
	List(ctx context.Context) ([]*types.Metrics, error)
}

// MetricListHTMLHandler handles HTTP requests to list all metrics in an HTML format.
type MetricListHTMLHandler struct {
	svc MetricLister
}

// MetricListHTMLHandlerOption defines a functional option for configuring MetricListHTMLHandler.
type MetricListHTMLHandlerOption func(*MetricListHTMLHandler)

// WithMetricLister sets the MetricLister service on MetricListHTMLHandler.
func WithMetricLister(svc MetricLister) MetricListHTMLHandlerOption {
	return func(h *MetricListHTMLHandler) {
		h.svc = svc
	}
}

// NewMetricListHTMLHandler creates a new MetricListHTMLHandler with the provided options.
func NewMetricListHTMLHandler(opts ...MetricListHTMLHandlerOption) *MetricListHTMLHandler {
	h := &MetricListHTMLHandler{}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// serveHTTP handles the HTTP request and responds with a HTML page listing all metrics.
func (h *MetricListHTMLHandler) serveHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	metrics, err := h.svc.List(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var builder strings.Builder
	builder.WriteString("<!DOCTYPE html><html><head><title>Metrics</title></head><body><ul>\n")
	for _, m := range metrics {
		switch m.Type {
		case types.Gauge:
			if m.Value != nil {
				builder.WriteString(fmt.Sprintf("<li>%s: %v</li>\n", m.ID, *m.Value))
			}
		case types.Counter:
			if m.Delta != nil {
				builder.WriteString(fmt.Sprintf("<li>%s: %d</li>\n", m.ID, *m.Delta))
			}
		}
	}
	builder.WriteString("</ul></body></html>\n")
	metricsHTML := builder.String()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(metricsHTML))
}

// RegisterRoute registers the route for serving the metrics list HTML page.
func (h *MetricListHTMLHandler) RegisterRoute(r chi.Router) {
	r.Get("/", h.serveHTTP)
}
