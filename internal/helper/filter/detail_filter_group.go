package filter

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// FilterGroup represents a group of filters that are combined with AND logic
type FilterGroup struct {
	Filters []*DetailFilter
}

type DetailFilterModel struct {
	Name       types.String `tfsdk:"name"`
	Value      types.String `tfsdk:"value"`
	Operator   types.String `tfsdk:"operator"`
	ValueRegex types.String `tfsdk:"value_regex"`
}

func NewFilterGroup(detailFilter []DetailFilterModel) (*FilterGroup, error) {
	if len(detailFilter) == 0 {
		return &FilterGroup{}, nil
	}

	filters := make([]*DetailFilter, 0, len(detailFilter))

	for _, filterModel := range detailFilter {
		filter, err := filterModel.newFilter()
		if err != nil {
			return nil, fmt.Errorf("error creating filter: %w", err)
		}
		filters = append(filters, filter)
	}

	return &FilterGroup{
		Filters: filters,
	}, nil
}

func (m *DetailFilterModel) newFilter() (*DetailFilter, error) {
	field := m.Name.ValueString()
	operator := m.Operator.ValueString()
	value := m.Value.ValueString()
	regex := m.ValueRegex.ValueString()

	return NewDetailFilter(field, operator, value, regex)
}

// Match determines if all filters in the group match the resource
// TODO: Implement OR logic if needed in the future
func (g *FilterGroup) Match(resource interface{}) bool {
	if g == nil || len(g.Filters) == 0 {
		return true
	}

	// All filters must match (AND logic)
	for _, filter := range g.Filters {
		if !filter.Match(resource) {
			return false
		}
	}

	return true
}
