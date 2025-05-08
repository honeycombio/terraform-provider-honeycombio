package models

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type BoardResourceModel struct {
	ID           types.String `tfsdk:"id"`
	BoardType    types.String `tfsdk:"type"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	ColumnLayout types.String `tfsdk:"column_layout"`
	Style        types.String `tfsdk:"style"`
	Queries      types.List   `tfsdk:"query"` // BoardQueryModel
	SLOs         types.Set    `tfsdk:"slo"`   // BoardSLOModel
	URL          types.String `tfsdk:"board_url"`
}

type BoardQueryModel struct {
	ID            types.String `tfsdk:"query_id"`
	AnnotationID  types.String `tfsdk:"query_annotation_id"`
	Caption       types.String `tfsdk:"caption"`
	Dataset       types.String `tfsdk:"dataset"`
	Style         types.String `tfsdk:"query_style"`
	GraphSettings types.List   `tfsdk:"graph_settings"` // BoardQueryGraphSettingsModel
}

var BoardQueryModelAttrType = map[string]attr.Type{
	"caption":             types.StringType,
	"dataset":             types.StringType,
	"query_id":            types.StringType,
	"query_annotation_id": types.StringType,
	"query_style":         types.StringType,
	"graph_settings": types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: BoardQueryGraphSettingsModelAttrType,
		},
	},
}

type BoardQueryGraphSettingsModel struct {
	LogScale          types.Bool `tfsdk:"log_scale"`
	OmitMissingValues types.Bool `tfsdk:"omit_missing_values"`
	HideMarkers       types.Bool `tfsdk:"hide_markers"`
	StackedGraphs     types.Bool `tfsdk:"stacked_graphs"`
	UTCXAxis          types.Bool `tfsdk:"utc_xaxis"`
	OverlaidCharts    types.Bool `tfsdk:"overlaid_charts"`
}

var BoardQueryGraphSettingsModelAttrType = map[string]attr.Type{
	"log_scale":           types.BoolType,
	"omit_missing_values": types.BoolType,
	"hide_markers":        types.BoolType,
	"stacked_graphs":      types.BoolType,
	"utc_xaxis":           types.BoolType,
	"overlaid_charts":     types.BoolType,
}

type BoardSLOModel struct {
	ID types.String `tfsdk:"id"`
}

var BoardSLOModelAttrType = map[string]attr.Type{
	"id": types.StringType,
}
