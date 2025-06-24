package services

import (
	"context"
	"sort"

	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// Getter defines an interface to get a metric by its MetricID.
type Getter interface {
	// Get fetches a metric by its ID and type.
	Get(ctx context.Context, id types.MetricID) (*types.Metrics, error)
}

// Saver defines an interface to save a metric.
type Saver interface {
	// Save stores or updates a metric.
	Save(ctx context.Context, metric types.Metrics) error
}

// Lister defines an interface to list all metrics.
type Lister interface {
	// List returns all stored metrics.
	List(ctx context.Context) ([]*types.Metrics, error)
}

// MetricUpdatesService provides methods to update metrics.
type MetricUpdatesService struct {
	getter Getter
	saver  Saver
}

// MetricUpdatesServiceOption defines a functional option for configuring MetricUpdatesService.
type MetricUpdatesServiceOption func(*MetricUpdatesService)

// WithMetricUpdatesGetter sets the Getter dependency for MetricUpdatesService.
func WithMetricUpdatesGetter(getter Getter) MetricUpdatesServiceOption {
	return func(svc *MetricUpdatesService) {
		svc.getter = getter
	}
}

// WithMetricUpdatesSaver sets the Saver dependency for MetricUpdatesService.
func WithMetricUpdatesSaver(saver Saver) MetricUpdatesServiceOption {
	return func(svc *MetricUpdatesService) {
		svc.saver = saver
	}
}

// NewMetricUpdatesService creates a new MetricUpdatesService with the provided options.
func NewMetricUpdatesService(opts ...MetricUpdatesServiceOption) *MetricUpdatesService {
	svc := &MetricUpdatesService{}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// Updates updates or adds the provided metrics, returning the updated metrics.
func (svc *MetricUpdatesService) Updates(
	ctx context.Context,
	metrics []*types.Metrics,
) ([]*types.Metrics, error) {
	metricsMap := make(map[types.MetricID]*types.Metrics)

	for _, m := range metrics {
		if m.Type == types.Counter {
			current, err := svc.getter.Get(ctx, types.MetricID{ID: m.ID, Type: m.Type})
			if err != nil {
				return nil, err
			}

			var currentDelta int64
			if current != nil && current.Delta != nil {
				currentDelta = *current.Delta
			}

			var newDelta int64
			if m.Delta != nil {
				newDelta = *m.Delta
			}

			totalDelta := currentDelta + newDelta
			m.Delta = &totalDelta
		}

		err := svc.saver.Save(ctx, *m)
		if err != nil {
			return nil, err
		}

		metricsMap[types.MetricID{ID: m.ID, Type: m.Type}] = m
	}

	updatedMetrics := make([]*types.Metrics, 0, len(metricsMap))
	for _, m := range metricsMap {
		updatedMetrics = append(updatedMetrics, m)
	}

	sort.Slice(updatedMetrics, func(i, j int) bool {
		return updatedMetrics[i].ID < updatedMetrics[j].ID
	})

	return updatedMetrics, nil
}

// MetricGetService provides method to get a single metric.
type MetricGetService struct {
	getter Getter
}

// MetricGetServiceOption defines a functional option for configuring MetricGetService.
type MetricGetServiceOption func(*MetricGetService)

// WithMetricGetGetter sets the Getter dependency for MetricGetService.
func WithMetricGetGetter(getter Getter) MetricGetServiceOption {
	return func(svc *MetricGetService) {
		svc.getter = getter
	}
}

// NewMetricGetService creates a new MetricGetService with the provided options.
func NewMetricGetService(opts ...MetricGetServiceOption) *MetricGetService {
	svc := &MetricGetService{}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// Get fetches a metric by its MetricID.
func (svc *MetricGetService) Get(
	ctx context.Context, metricID types.MetricID,
) (*types.Metrics, error) {
	return svc.getter.Get(ctx, metricID)
}

// MetricListService provides method to list all metrics.
type MetricListService struct {
	lister Lister
}

// MetricListServiceOption defines a functional option for configuring MetricListService.
type MetricListServiceOption func(*MetricListService)

// WithMetricListLister sets the Lister dependency for MetricListService.
func WithMetricListLister(lister Lister) MetricListServiceOption {
	return func(svc *MetricListService) {
		svc.lister = lister
	}
}

// NewMetricListService creates a new MetricListService with the provided options.
func NewMetricListService(opts ...MetricListServiceOption) *MetricListService {
	svc := &MetricListService{}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// List returns all stored metrics.
func (svc *MetricListService) List(ctx context.Context) ([]*types.Metrics, error) {
	return svc.lister.List(ctx)
}
