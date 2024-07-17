package honeycombio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_coerceValueToType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name:     "boolean true",
			input:    "true",
			expected: true,
		},
		{
			name:     "boolean false",
			input:    "false",
			expected: false,
		},
		{
			name:     "float",
			input:    "10383.383",
			expected: 10383.383,
		},
		{
			name:     "int",
			input:    "300",
			expected: int64(300),
		},
		{
			name:     "zero",
			input:    "0",
			expected: int64(0),
		},
		{
			name:     "stringy number",
			input:    "10.special",
			expected: "10.special",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, coerceValueToType(tt.input))
		})
	}
}
