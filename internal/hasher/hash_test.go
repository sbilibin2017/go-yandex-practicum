package hasher

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHash(t *testing.T) {
	key := "secret"
	data := []byte("important data")

	hash1 := Hash(data, key)
	require.NotEmpty(t, hash1, "Hash should not be empty")

	// Повторное вычисление должно дать тот же результат
	hash2 := Hash(data, key)
	assert.Equal(t, hash1, hash2, "Hashes should match for same input and key")

	// С другим ключом — другой результат
	differentKeyHash := Hash(data, "otherkey")
	assert.NotEqual(t, hash1, differentKeyHash, "Hashes should not match for different keys")
}

func TestCompare(t *testing.T) {
	hash1 := "abc123"
	hash2 := "abc123"
	hash3 := "def456"

	assert.True(t, Compare(hash1, hash2), "Hashes should match")
	assert.False(t, Compare(hash1, hash3), "Hashes should not match")
}
