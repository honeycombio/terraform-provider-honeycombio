package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
				mustCreateFilter(t, "ID", "starts-with", "abc", ""),
				mustCreateFilter(t, "Count", ">", "30", ""),
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
			got := group.Match(resource)

			assert.Equal(t, tt.want, got, "FilterGroup.Match() mismatch for test case: %s", tt.name)
		})
	}
}

// Helper function to create a filter without having to check errors in each test case
func mustCreateFilter(t *testing.T, field, operator, value, regex string) *DetailFilter {
	filter, err := NewDetailFilter(field, operator, value, regex)
	assert.NoError(t, err)
	return filter
}
