package honeycombio

import (
	"reflect"
	"testing"
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
			expected: 300,
		},
		{
			name:     "zero",
			input:    "0",
			expected: 0,
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
			if got := coerceValueToType(tt.input); !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("coerceInputToType() = %v<%T>, want %v<%T>", got, got, tt.expected, tt.expected)
			}
		})
	}
}
