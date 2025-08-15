package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestResource is a struct for testing field matching
type TestResource struct {
	Name        string
	ID          string
	Count       int
	Price       float64
	IsAvailable bool
	Tags        map[string]string
	Scores      []int
	Properties  map[string]any
}

func TestMatch(t *testing.T) {
	resource := TestResource{
		Name:        "Test Resource",
		ID:          "123456",
		Count:       42,
		Price:       99.99,
		IsAvailable: true,
		Tags: map[string]string{
			"environment": "production",
			"owner":       "team-a",
			"region":      "us-west",
		},
		Scores: []int{85, 90, 95},
		Properties: map[string]any{
			"nested": map[string]string{
				"key":        "value",
				"nested_key": "nested_value",
			},
			"version":    "1.0",
			"created_at": "2023-10-01T12:00:00Z",
			"updated_at": "2023-10-02T12:00:00Z",
		},
	}

	tests := []struct {
		name     string
		field    string
		operator string
		value    string
		regex    string
		expected bool
		useMap   bool // whether to use the map resource instead of struct
	}{
		// String field tests
		{
			name:     "String equals - match",
			field:    "Name",
			operator: "equals",
			value:    "Test Resource",
			expected: true,
		},
		{
			name:     "String equals - no match",
			field:    "Name",
			operator: "equals",
			value:    "Different",
			expected: false,
		},
		{
			name:     "String not_equals - match",
			field:    "Name",
			operator: "not-equals",
			value:    "Different",
			expected: true,
		},
		{
			name:     "String contains - match",
			field:    "Name",
			operator: "contains",
			value:    "Resource",
			expected: true,
		},
		{
			name:     "No operator - match",
			field:    "Name",
			value:    "Test Resource",
			expected: true,
		},
		{
			name:     "String starts_with - match",
			field:    "Name",
			operator: "starts-with",
			value:    "Test",
			expected: true,
		},
		{
			name:     "String ends_with - match",
			field:    "Name",
			operator: "ends-with",
			value:    "Resource",
			expected: true,
		},
		{
			name:     "String regex - match",
			field:    "Name",
			regex:    "Test.*",
			expected: true,
		},

		// Integer field tests
		{
			name:     "Integer equals - match",
			field:    "Count",
			operator: "equals",
			value:    "42",
			expected: true,
		},
		{
			name:     "Integer greater than - match",
			field:    "Count",
			operator: ">",
			value:    "30",
			expected: true,
		},
		{
			name:     "Integer less than - match",
			field:    "Count",
			operator: "lt",
			value:    "50",
			expected: true,
		},

		// Float field tests
		{
			name:     "Float equals - match",
			field:    "Price",
			operator: "equals",
			value:    "99.990000", // Note the format from fmt.Sprintf("%f", ...)
			expected: true,
		},
		{
			name:     "Float greater than - match",
			field:    "Price",
			operator: ">",
			value:    "50.5",
			expected: true,
		},

		// Boolean field tests
		{
			name:     "Bool equals - match",
			field:    "IsAvailable",
			operator: "equals",
			value:    "true",
			expected: true,
		},
		{
			name:     "Bool equals - match",
			field:    "is_available",
			operator: "equals",
			value:    "true",
			expected: true,
		},

		// Edge cases
		{
			name:     "Non-existent field",
			field:    "NonExistentField",
			operator: "equals",
			value:    "anything",
			expected: false,
		},

		// Map field tests (using struct with map field)
		{
			name:     "Map field contains check",
			field:    "Tags",
			operator: "contains",
			value:    "production",
			expected: true,
		},

		// List field tests
		{
			name:     "List field contains check",
			field:    "Scores",
			operator: "contains",
			value:    "90",
			expected: true,
		},

		// Nested map fields
		{
			name:     "Nested map in Properties",
			field:    "Properties",
			operator: "contains",
			value:    "nested",
			expected: true,
		},

		// New operators test cases
		{
			name:     "String does-not-start-with - match",
			field:    "Name",
			operator: "does-not-start-with",
			value:    "Wrong",
			expected: true,
		},
		{
			name:     "String does-not-start-with - no match",
			field:    "Name",
			operator: "does-not-start-with",
			value:    "Test",
			expected: false,
		},
		{
			name:     "String does-not-end-with - match",
			field:    "Name",
			operator: "does-not-end-with",
			value:    "Wrong",
			expected: true,
		},
		{
			name:     "String does-not-end-with - no match",
			field:    "Name",
			operator: "does-not-end-with",
			value:    "Resource",
			expected: false,
		},
		{
			name:     "String does-not-contain - match",
			field:    "Name",
			operator: "does-not-contain",
			value:    "Wrong",
			expected: true,
		},
		{
			name:     "String does-not-contain - no match",
			field:    "Name",
			operator: "does-not-contain",
			value:    "Resource",
			expected: false,
		},
		{
			name:     "String not-in - match",
			field:    "Name",
			operator: "not-in",
			value:    "Wrong",
			expected: true,
		},
		{
			name:     "String not-in - no match",
			field:    "Name",
			operator: "not-in",
			value:    "Resource",
			expected: false,
		},
		{
			name:     "Integer greater_than_or_equal - match equal",
			field:    "Count",
			operator: ">=",
			value:    "42",
			expected: true,
		},
		{
			name:     "Integer greater_than_or_equal - match greater",
			field:    "Count",
			operator: ">=",
			value:    "30",
			expected: true,
		},
		{
			name:     "Integer greater_than_or_equal - no match",
			field:    "Count",
			operator: ">=",
			value:    "50",
			expected: false,
		},
		{
			name:     "Integer less_than_or_equal - match equal",
			field:    "Count",
			operator: "<=",
			value:    "42",
			expected: true,
		},
		{
			name:     "Integer less_than_or_equal - match less",
			field:    "Count",
			operator: "<=",
			value:    "50",
			expected: true,
		},
		{
			name:     "Integer less_than_or_equal - no match",
			field:    "Count",
			operator: "<=",
			value:    "30",
			expected: false,
		},
		{
			name:     "String does-not-exist - match for empty string",
			field:    "Name",
			operator: "does-not-exist",
			value:    "",
			expected: false,
		},
		{
			name:     "NonExistentField does-not-exist - no match",
			field:    "NonExistentField",
			operator: "does-not-exist",
			value:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := NewDetailFilter(tt.field, tt.operator, tt.value, tt.regex)
			if err != nil {
				t.Fatalf("Failed to create filter: %v", err)
			}
			result := filter.Match(resource)

			assert.Equal(t, tt.expected, result, "Filter match result mismatch for %s", tt.name)
		})
	}
}
