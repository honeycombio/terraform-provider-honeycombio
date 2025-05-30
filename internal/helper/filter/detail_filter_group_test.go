package filter

import (
	"testing"
)

// TestResource for testing filter matching
type TestFilterResource struct {
	Name        string
	ID          string
	Description string
	Count       int
	Active      bool
	Tags        map[string]string
}

func TestNewFilterGroup(t *testing.T) {
	// Test with empty filters
	emptyFilters := []*DetailFilter{}
	group := NewFilterGroup(emptyFilters)
	if group == nil {
		t.Fatal("Expected non-nil group for empty filters")
	}
	if len(group.Filters) != 0 {
		t.Errorf("Expected 0 filters, got %d", len(group.Filters))
	}

	// Test with non-empty filters
	filter1, _ := NewDetailFilter("name", "equals", "test", "")
	filter2, _ := NewDetailFilter("id", "contains", "abc", "")
	filters := []*DetailFilter{filter1, filter2}

	group = NewFilterGroup(filters)
	if group == nil {
		t.Fatal("Expected non-nil group")
	}
	if len(group.Filters) != 2 {
		t.Errorf("Expected 2 filters, got %d", len(group.Filters))
	}
}

func TestFilterGroup_Match(t *testing.T) {
	resource := TestFilterResource{
		Name:        "Test Resource",
		ID:          "abc123",
		Description: "This is a test resource",
		Count:       42,
		Active:      true,
		Tags: map[string]string{
			"env": "test",
			"app": "example",
		},
	}

	tests := []struct {
		name    string
		filters []*DetailFilter
		want    bool
	}{
		{
			name:    "nil filters",
			filters: nil,
			want:    true,
		},
		{
			name:    "empty filters",
			filters: []*DetailFilter{},
			want:    true,
		},
		{
			name: "single matching filter",
			filters: []*DetailFilter{
				mustCreateFilter(t, "Name", "equals", "Test Resource", ""),
			},
			want: true,
		},
		{
			name: "single non-matching filter",
			filters: []*DetailFilter{
				mustCreateFilter(t, "Name", "equals", "Wrong Name", ""),
			},
			want: false,
		},
		{
			name: "multiple matching filters",
			filters: []*DetailFilter{
				mustCreateFilter(t, "Name", "contains", "Resource", ""),
				mustCreateFilter(t, "ID", "starts_with", "abc", ""),
				mustCreateFilter(t, "Count", "greater_than", "30", ""),
			},
			want: true,
		},
		{
			name: "mixed matching and non-matching filters",
			filters: []*DetailFilter{
				mustCreateFilter(t, "Name", "contains", "Resource", ""),
				mustCreateFilter(t, "ID", "equals", "wrong-id", ""), // This won't match
			},
			want: false,
		},
		{
			name: "non-existent field filter",
			filters: []*DetailFilter{
				mustCreateFilter(t, "NonExistentField", "equals", "anything", ""),
			},
			want: false,
		},
		{
			name: "regex filter",
			filters: []*DetailFilter{
				mustCreateFilter(t, "Name", "", "", "Test.*"),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := NewFilterGroup(tt.filters)
			if got := group.Match(resource); got != tt.want {
				t.Errorf("FilterGroup.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterGroup_MatchName(t *testing.T) {
	tests := []struct {
		name     string
		filters  []*DetailFilter
		testName string
		want     bool
	}{
		{
			name:     "nil filters",
			filters:  nil,
			testName: "any name",
			want:     true,
		},
		{
			name:     "empty filters",
			filters:  []*DetailFilter{},
			testName: "any name",
			want:     true,
		},
		{
			name: "single matching name filter",
			filters: []*DetailFilter{
				mustCreateFilter(t, "name", "equals", "test-name", ""),
			},
			testName: "test-name",
			want:     true,
		},
		{
			name: "single non-matching name filter",
			filters: []*DetailFilter{
				mustCreateFilter(t, "name", "equals", "test-name", ""),
			},
			testName: "different-name",
			want:     false,
		},
		{
			name: "mixed matching and non-matching name filters",
			filters: []*DetailFilter{
				mustCreateFilter(t, "name", "starts_with", "test", ""),
				mustCreateFilter(t, "name", "equals", "wrong-name", ""), // This won't match
			},
			testName: "test-name",
			want:     false,
		},
		{
			name: "regex name filter",
			filters: []*DetailFilter{
				mustCreateFilter(t, "name", "", "", "test-.*"),
			},
			testName: "test-123",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := NewFilterGroup(tt.filters)
			if got := group.MatchName(tt.testName); got != tt.want {
				t.Errorf("FilterGroup.MatchName() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to create a filter without having to check errors in each test case
func mustCreateFilter(t *testing.T, field, operator, value, regex string) *DetailFilter {
	filter, err := NewDetailFilter(field, operator, value, regex)
	if err != nil {
		t.Fatalf("Failed to create filter: %v", err)
	}
	return filter
}
