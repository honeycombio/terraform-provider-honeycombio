package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/filter"
)

type DetailFilterModel struct {
	Name       types.String `tfsdk:"name"`
	Value      types.String `tfsdk:"value"`
	ValueRegex types.String `tfsdk:"value_regex"`
}

func (f *DetailFilterModel) NewFilter() (*filter.DetailFilter, error) {
	if f == nil {
		return nil, nil
	}
	return filter.NewDetailFilter(f.Name.ValueString(), f.Value.ValueString(), f.ValueRegex.ValueString())
}
