package filter

import (
	"testing"
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
	Properties  map[string]interface{}
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
		Properties: map[string]interface{}{
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
			operator: "not_equals",
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
			name:     "String starts_with - match",
			field:    "Name",
			operator: "starts_with",
			value:    "Test",
			expected: true,
		},
		{
			name:     "String ends_with - match",
			field:    "Name",
			operator: "ends_with",
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
			name:     "Integer greater_than - match",
			field:    "Count",
			operator: "greater_than",
			value:    "30",
			expected: true,
		},
		{
			name:     "Integer less_than - match",
			field:    "Count",
			operator: "less_than",
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
			name:     "Float greater_than - match",
			field:    "Price",
			operator: "greater_than",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := NewDetailFilter(tt.field, tt.operator, tt.value, tt.regex)
			if err != nil {
				t.Fatalf("Failed to create filter: %v", err)
			}
			result := filter.Match(resource)

			if result != tt.expected {
				t.Errorf("Expected Match to return %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMatchName(t *testing.T) {
	tests := []struct {
		name           string
		filterField    string
		filterOperator string
		filterValue    string
		filterRegex    string
		testName       string
		expected       bool
	}{
		{
			name:           "Match name with equals",
			filterField:    "name",
			filterOperator: "equals",
			filterValue:    "test-name",
			testName:       "test-name",
			expected:       true,
		},
		{
			name:           "No match name with equals",
			filterField:    "name",
			filterOperator: "equals",
			filterValue:    "test-name",
			testName:       "different-name",
			expected:       false,
		},
		{
			name:        "Match name with regex",
			filterField: "name",
			filterRegex: "test-.*",
			testName:    "test-123",
			expected:    true,
		},
		{
			name:           "Different field should always match name",
			filterField:    "id",
			filterOperator: "equals",
			filterValue:    "123",
			testName:       "any-name",
			expected:       true,
		},
		{
			name:     "Nil filter should match any name",
			testName: "any-name",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filter *DetailFilter
			if tt.filterField != "" || tt.filterOperator != "" || tt.filterValue != "" || tt.filterRegex != "" {
				var err error
				filter, err = NewDetailFilter(tt.filterField, tt.filterOperator, tt.filterValue, tt.filterRegex)
				if err != nil {
					t.Fatalf("Failed to create filter: %v", err)
				}
			}

			result := filter.MatchName(tt.testName)
			if result != tt.expected {
				t.Errorf("Expected MatchName to return %v, got %v", tt.expected, result)
			}
		})
	}
}
