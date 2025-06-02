package models

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/filter"
)

type DetailFilterModel struct {
	Name       types.String `tfsdk:"name"`
	Value      types.String `tfsdk:"value"`
	Operator   types.String `tfsdk:"operator"`
	ValueRegex types.String `tfsdk:"value_regex"`
}

// NewFilterGroup creates a filter group from multiple filter models
func NewFilterGroup(detailFilter []DetailFilterModel) (*filter.FilterGroup, error) {
	if len(detailFilter) == 0 {
		return nil, nil
	}

	filters := make([]*filter.DetailFilter, 0, len(detailFilter))

	for _, filterModel := range detailFilter {
		filter, err := filterModel.newFilter()
		if err != nil {
			return nil, fmt.Errorf("error creating filter: %w", err)
		}
		filters = append(filters, filter)
	}

	return filter.NewFilterGroup(filters), nil
}

// newFilter creates a new filter from the filter model
func (m *DetailFilterModel) newFilter() (*filter.DetailFilter, error) {
	field := m.Name.ValueString()
	operator := m.Operator.ValueString()
	value := m.Value.ValueString()
	regex := m.ValueRegex.ValueString()

	return filter.NewDetailFilter(field, operator, value, regex)
}
