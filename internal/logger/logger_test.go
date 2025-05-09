package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitialize_Initialize(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		expectError   bool
		expectLogInit bool
	}{
		{
			name:          "Valid level - info",
			level:         "info",
			expectError:   false,
			expectLogInit: true,
		},
		{
			name:          "Valid level - debug",
			level:         "debug",
			expectError:   false,
			expectLogInit: true,
		},
		{
			name:          "Invalid level",
			level:         "notalevel",
			expectError:   true,
			expectLogInit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetLogger()

			err := Initialize(tt.level)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, Log)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, Log)
			}
		})
	}
}

func TestInitialize_OnceOnly(t *testing.T) {
	resetLogger()

	err1 := Initialize("debug")
	assert.NoError(t, err1)
	firstInstance := Log

	err2 := Initialize("warn") // должно быть проигнорировано
	assert.NoError(t, err2)
	secondInstance := Log

	assert.Equal(t, firstInstance, secondInstance)
}

func resetLogger() {
	Log = nil
}
