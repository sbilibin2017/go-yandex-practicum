package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashWithKey(t *testing.T) {
	key := "secret"
	data := []byte("hello world")
	hash1 := HashWithKey(data, key)
	hash2 := HashWithKey(data, key)
	assert.NotEmpty(t, hash1)
	assert.Equal(t, hash1, hash2, "Hashes for the same data and key must be equal")
	hash3 := HashWithKey([]byte("hello"), key)
	hash4 := HashWithKey(data, "different_key")
	assert.NotEqual(t, hash1, hash3)
	assert.NotEqual(t, hash1, hash4)
}

func TestCompareHash(t *testing.T) {
	key := "secret"
	data := []byte("test data")
	hash1 := HashWithKey(data, key)
	hash2 := HashWithKey(data, key)
	hash3 := HashWithKey([]byte("other data"), key)
	assert.True(t, CompareHash(hash1, hash2))
	assert.False(t, CompareHash(hash1, hash3))
	assert.False(t, CompareHash(hash1, ""))
	assert.False(t, CompareHash("", hash2))
	assert.True(t, CompareHash("", ""))
}
