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

func TestSync(t *testing.T) {
	err := Sync()
	assert.NoError(t, err)

	// Case 2: Log is a valid sugared logger
	err = Initialize("info")
	assert.NoError(t, err)

	err = Sync()
	assert.NoError(t, err)
}

func resetLogger() {
	Log = nil
}
