package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExit(t *testing.T) {
	t.Run("returns 1 when error is not nil", func(t *testing.T) {
		err := assert.AnError
		code := exit(err)
		assert.Equal(t, 1, code)
	})

	t.Run("returns 0 when error is nil", func(t *testing.T) {
		code := exit(nil)
		assert.Equal(t, 0, code)
	})
}
