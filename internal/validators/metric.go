package validators

import (
	"strconv"

	"github.com/sbilibin2017/go-yandex-practicum/internal/errors"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

func ValidateMetricID(id types.MetricID) error {
	if id.ID == "" {
		return errors.ErrMetricIDRequired
	}
	if id.MType != types.Counter && id.MType != types.Gauge {
		return errors.ErrUnsupportedMetricType

	}
	return nil
}

func ValidateMetric(metrics *types.Metrics) error {
	err := ValidateMetricID(types.MetricID{ID: metrics.ID, MType: metrics.MType})
	if err != nil {
		return err
	}

	switch metrics.MType {
	case types.Counter:
		if metrics.Delta == nil {
			return errors.ErrCounterValueRequired
		}
	case types.Gauge:
		if metrics.Value == nil {
			return errors.ErrGaugeValueRequired
		}
	}

	return nil
}

func ValidateMetricIDAttributes(mType string, mID string) error {
	if mID == "" {
		return errors.ErrMetricNameRequired
	}

	if mType != types.Counter && mType != types.Gauge {
		return errors.ErrUnsupportedMetricType
	}

	return nil
}

func ValidateMetricAttributes(mType string, mName string, mValue string) error {
	err := ValidateMetricIDAttributes(mType, mName)
	if err != nil {
		return err
	}

	switch mType {
	case types.Counter:
		if _, err := strconv.ParseInt(mValue, 10, 64); err != nil {
			return errors.ErrInvalidCounterValue
		}
	case types.Gauge:
		if _, err := strconv.ParseFloat(mValue, 64); err != nil {
			return errors.ErrInvalidGaugeValue
		}
	}

	return nil
}
