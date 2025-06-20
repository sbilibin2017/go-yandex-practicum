package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintBuildInfo(t *testing.T) {
	// Backup original values and restore after test
	origVersion := buildVersion
	origDate := buildDate
	origCommit := buildCommit
	defer func() {
		buildVersion = origVersion
		buildDate = origDate
		buildCommit = origCommit
	}()

	tests := []struct {
		name     string
		version  string
		date     string
		commit   string
		expected []string
	}{
		{
			name:     "All values set",
			version:  "1.2.3",
			date:     "2025-06-19",
			commit:   "abc123",
			expected: []string{"Build version: 1.2.3", "Build date: 2025-06-19", "Build commit: abc123"},
		},
		{
			name:     "Empty values",
			version:  "",
			date:     "",
			commit:   "",
			expected: []string{"Build version: N/A", "Build date: N/A", "Build commit: N/A"},
		},
		{
			name:     "Some values empty",
			version:  "2.0.0",
			date:     "",
			commit:   "def456",
			expected: []string{"Build version: 2.0.0", "Build date: N/A", "Build commit: def456"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buildVersion = tt.version
			buildDate = tt.date
			buildCommit = tt.commit

			// Redirect stdout
			r, w, err := os.Pipe()
			assert.NoError(t, err)

			stdout := os.Stdout
			os.Stdout = w

			// Run the function
			printBuildInfo()

			// Close writer and restore stdout
			w.Close()
			os.Stdout = stdout

			var buf bytes.Buffer
			_, err = buf.ReadFrom(r)
			assert.NoError(t, err)
			r.Close()

			output := buf.String()
			for _, exp := range tt.expected {
				assert.Contains(t, output, exp)
			}
		})
	}
}
