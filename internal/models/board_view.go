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
	Column    types.String   `tfsdk:"column"`
	Operation types.String   `tfsdk:"operation"`
	Value     types.Dynamic  `tfsdk:"value"`
}

var BoardViewFilterModelAttrType = map[string]attr.Type{
	"column":    types.StringType,
	"operation": types.StringType,
	"value":     types.DynamicType,
}

