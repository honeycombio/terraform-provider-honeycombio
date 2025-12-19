package models

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type BoardViewResourceModel struct {
	ID      types.String `tfsdk:"id"`
	BoardID types.String `tfsdk:"board_id"`
	Name    types.String `tfsdk:"name"`
	Filters types.List   `tfsdk:"filter"`
}

type BoardViewFilterModel struct {
	Column    types.String `tfsdk:"column"`
	Operation types.String `tfsdk:"operation"`
	Value     types.String `tfsdk:"value"` // TODO: convert to types.DynamicType when supported in nested blocks
}

var BoardViewFilterModelAttrType = map[string]attr.Type{
	"column":    types.StringType,
	"operation": types.StringType,
	"value":     types.StringType,
}

