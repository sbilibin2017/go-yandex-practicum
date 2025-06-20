package validators

import (
	"testing"

	"github.com/sbilibin2017/go-yandex-practicum/internal/errors"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestValidateMetricID(t *testing.T) {
	tests := []struct {
		name      string
		input     types.MetricID
		wantError error
	}{
		{"valid counter", types.MetricID{ID: "id1", MType: types.Counter}, nil},
		{"valid gauge", types.MetricID{ID: "id2", MType: types.Gauge}, nil},
		{"empty id", types.MetricID{ID: "", MType: types.Counter}, errors.ErrMetricIDRequired},
		{"unsupported type", types.MetricID{ID: "id3", MType: "unknown"}, errors.ErrUnsupportedMetricType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMetricID(tt.input)
			assert.ErrorIs(t, err, tt.wantError)
		})
	}
}

func TestValidateMetric(t *testing.T) {
	deltaVal := int64(10)
	valueVal := float64(1.23)

	tests := []struct {
		name      string
		input     *types.Metrics
		wantError error
	}{
		{"valid counter", &types.Metrics{ID: "id1", MType: types.Counter, Delta: &deltaVal}, nil},
		{"valid gauge", &types.Metrics{ID: "id2", MType: types.Gauge, Value: &valueVal}, nil},
		{"counter missing delta", &types.Metrics{ID: "id3", MType: types.Counter, Delta: nil}, errors.ErrCounterValueRequired}, // fixed here
		{"gauge missing value", &types.Metrics{ID: "id4", MType: types.Gauge, Value: nil}, errors.ErrGaugeValueRequired},       // fixed here
		{"invalid id", &types.Metrics{ID: "", MType: types.Counter, Delta: &deltaVal}, errors.ErrMetricIDRequired},
		{"unsupported type", &types.Metrics{ID: "id5", MType: "unknown"}, errors.ErrUnsupportedMetricType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMetric(tt.input)
			assert.ErrorIs(t, err, tt.wantError)
		})
	}
}

func TestValidateMetricIDAttributes(t *testing.T) {
	tests := []struct {
		name      string
		mType     string
		mID       string
		wantError error
	}{
		{"valid counter", types.Counter, "id1", nil},
		{"valid gauge", types.Gauge, "id2", nil},
		{"empty id", types.Counter, "", errors.ErrMetricNameRequired},
		{"unsupported type", "unknown", "id3", errors.ErrUnsupportedMetricType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMetricIDAttributes(tt.mType, tt.mID)
			assert.ErrorIs(t, err, tt.wantError)
		})
	}
}

func TestValidateMetricAttributes(t *testing.T) {
	tests := []struct {
		name      string
		mType     string
		mName     string
		mValue    string
		wantError error
	}{
		{"valid counter", types.Counter, "id1", "123", nil},
		{"valid gauge", types.Gauge, "id2", "3.1415", nil},
		{"invalid counter value", types.Counter, "id3", "abc", errors.ErrInvalidCounterValue},
		{"invalid gauge value", types.Gauge, "id4", "notafloat", errors.ErrInvalidGaugeValue},
		{"empty name", types.Counter, "", "123", errors.ErrMetricNameRequired},
		{"unsupported type", "unknown", "id5", "123", errors.ErrUnsupportedMetricType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMetricAttributes(tt.mType, tt.mName, tt.mValue)
			assert.ErrorIs(t, err, tt.wantError)
		})
	}
}
