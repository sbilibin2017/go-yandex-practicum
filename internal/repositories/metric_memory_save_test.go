package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricMemorySaveRepository_Save(t *testing.T) {
	initialData := make(map[string]any)
	repo := NewMetricMemorySaveRepository(initialData)
	ctx := context.Background()
	data := map[string]any{
		"id":    "metric1",
		"type":  "counter",
		"delta": int64(10),
	}
	err := repo.Save(ctx, data)
	assert.NoError(t, err)
	key := generateMetricKey(data)
	storedData, exists := repo.data[key]
	assert.True(t, exists, "Data should be saved in the repository")
	assert.Equal(t, data, storedData, "Saved data should match the input data")
}
