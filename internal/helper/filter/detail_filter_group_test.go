package filter

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	emptyFilters := []DetailFilterModel{}
	group, err := NewFilterGroup(emptyFilters)
	require.NoError(t, err)
	assert.NotNil(t, group, "Expected non-nil group for empty filters")
	assert.Empty(t, group.Filters, "Expected 0 filters for empty input")

	// Test with non-empty filters
	filter1 := DetailFilterModel{
		Name:     types.StringValue("name"),
		Operator: types.StringValue("equals"),
		Value:    types.StringValue("test"),
	}
	filter2 := DetailFilterModel{
		Name:     types.StringValue("id"),
		Operator: types.StringValue("contains"),
		Value:    types.StringValue("abc"),
	}
	filters := []DetailFilterModel{filter1, filter2}

	group, err = NewFilterGroup(filters)
	require.NoError(t, err)
	assert.NotNil(t, group, "Expected non-nil group for non-empty filters")
	assert.Len(t, group.Filters, 2, "Expected 2 filters for non-empty input")
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
		filters []DetailFilterModel
		want    bool
	}{
		{
			name:    "nil filters",
			filters: nil,
			want:    true,
		},
		{
			name:    "empty filters",
			filters: []DetailFilterModel{},
			want:    true,
		},
		{
			name: "single matching filter",
			filters: []DetailFilterModel{
				mustCreateFilter(t, "Name", "equals", "Test Resource", ""),
			},
			want: true,
		},
		{
			name: "single non-matching filter",
			filters: []DetailFilterModel{
				mustCreateFilter(t, "Name", "equals", "Wrong Name", ""),
			},
			want: false,
		},
		{
			name: "multiple matching filters",
			filters: []DetailFilterModel{
				mustCreateFilter(t, "Name", "contains", "Resource", ""),
				mustCreateFilter(t, "ID", "starts-with", "abc", ""),
				mustCreateFilter(t, "Count", ">", "30", ""),
			},
			want: true,
		},
		{
			name: "mixed matching and non-matching filters",
			filters: []DetailFilterModel{
				mustCreateFilter(t, "Name", "contains", "Resource", ""),
				mustCreateFilter(t, "ID", "equals", "wrong-id", ""), // This won't match
			},
			want: false,
		},
		{
			name: "non-existent field filter",
			filters: []DetailFilterModel{
				mustCreateFilter(t, "NonExistentField", "equals", "anything", ""),
			},
			want: false,
		},
		{
			name: "regex filter",
			filters: []DetailFilterModel{
				mustCreateFilter(t, "Name", "", "", "Test.*"),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, err := NewFilterGroup(tt.filters)
			require.NoError(t, err)
			got := group.Match(resource)

			assert.Equal(t, tt.want, got, "FilterGroup.Match() mismatch for test case: %s", tt.name)
		})
	}
}

// Helper function to create a filter without having to check errors in each test case
func mustCreateFilter(_ *testing.T, field, operator, value, regex string) DetailFilterModel {
	return DetailFilterModel{
		Name:       types.StringValue(field),
		Operator:   types.StringValue(operator),
		Value:      types.StringValue(value),
		ValueRegex: types.StringValue(regex),
	}

}
