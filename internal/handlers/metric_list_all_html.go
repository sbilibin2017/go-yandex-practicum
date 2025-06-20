package http

import (
	"context"
	"net/http"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

type MetricListAllHTMLService interface {
	ListAll(ctx context.Context) ([]*types.Metrics, error)
}

type MetricListAllHTMLHandler struct {
	svc MetricListAllHTMLService
}

func NewMetricListAllHTMLHandler(
	svc MetricListAllHTMLService,
) *MetricListAllHTMLHandler {
	return &MetricListAllHTMLHandler{svc: svc}
}

func (h *MetricListAllHTMLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.svc.ListAll(r.Context())
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(types.NewMetricsHTML(metrics)))
}
