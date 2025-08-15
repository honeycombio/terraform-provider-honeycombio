package coerce_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/coerce"
)

func Test_valueToType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected any
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
			assert.Equal(t, tt.expected, coerce.ValueToType(tt.input))
		})
	}
}

func Test_valueToString(t *testing.T) {
	type testStruct struct {
		Name  string
		Value int
	}

	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "string",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "int",
			input:    42,
			expected: "42",
		},
		{
			name:     "int64",
			input:    int64(9223372036854775807),
			expected: "9223372036854775807",
		},
		{
			name:     "float32",
			input:    float32(3.14159),
			expected: "3.141590",
		},
		{
			name:     "float64",
			input:    float64(3.14159265359),
			expected: "3.141593",
		},
		{
			name:     "boolean true",
			input:    true,
			expected: "true",
		},
		{
			name:     "boolean false",
			input:    false,
			expected: "false",
		},
		{
			name:     "slice",
			input:    []int{1, 2, 3},
			expected: "[1 2 3]",
		},
		{
			name:     "map",
			input:    map[string]int{"one": 1, "two": 2},
			expected: "map[one:1 two:2]",
		},
		{
			name:     "struct",
			input:    testStruct{Name: "test", Value: 42},
			expected: "{test 42}",
		},
		{
			name:     "terraform string value",
			input:    basetypes.NewStringValue("test.6UROptj5Vt"),
			expected: "test.6UROptj5Vt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := coerce.ValueToString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
