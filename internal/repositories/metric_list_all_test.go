package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricListAllRepository_ListAll(t *testing.T) {
	ctx := context.Background()

	t.Run("returns sorted list of metrics", func(t *testing.T) {
		repo := NewMetricListAllRepository(map[string]any{
			"2": map[string]any{"id": "2", "type": "gauge", "value": 3.14},
			"1": map[string]any{"id": "1", "type": "counter", "delta": 10},
			"3": map[string]any{"id": "3", "type": "gauge", "value": 1.23},
		})

		result, err := repo.ListAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, "1", result[0]["id"])
		assert.Equal(t, "2", result[1]["id"])
		assert.Equal(t, "3", result[2]["id"])
	})

	t.Run("returns nil when no valid metrics found", func(t *testing.T) {
		repo := NewMetricListAllRepository(map[string]any{
			"invalid": "not a map",
		})

		result, err := repo.ListAll(ctx)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("returns empty when data is empty", func(t *testing.T) {
		repo := NewMetricListAllRepository(map[string]any{})

		result, err := repo.ListAll(ctx)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}
