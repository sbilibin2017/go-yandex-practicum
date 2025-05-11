package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricFilterOneRepository_FilterOne_Found(t *testing.T) {
	initialData := map[string]any{
		"metric1:counter": map[string]any{
			"id":    "metric1",
			"type":  "counter",
			"delta": int64(10),
		},
	}
	repo := NewMetricFilterOneRepository(initialData)
	ctx := context.Background()
	filter := map[string]any{
		"id":   "metric1",
		"type": "counter",
	}
	result, err := repo.FilterOne(ctx, filter)
	assert.NoError(t, err)
	expected := map[string]any{
		"id":    "metric1",
		"type":  "counter",
		"delta": int64(10),
	}
	assert.Equal(t, expected, result)
}

func TestMetricFilterOneRepository_FilterOne_NotFound(t *testing.T) {
	initialData := map[string]any{}
	repo := NewMetricFilterOneRepository(initialData)
	ctx := context.Background()
	filter := map[string]any{
		"id":   "metric1",
		"type": "counter",
	}
	result, err := repo.FilterOne(ctx, filter)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestMetricFilterOneRepository_FilterOne_EmptyData(t *testing.T) {
	initialData := map[string]any{}
	repo := NewMetricFilterOneRepository(initialData)
	ctx := context.Background()
	filter := map[string]any{
		"id":   "metric1",
		"type": "counter",
	}
	result, err := repo.FilterOne(ctx, filter)
	assert.NoError(t, err)
	assert.Nil(t, result)
}
