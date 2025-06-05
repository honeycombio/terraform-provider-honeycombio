package models

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type FlexibleBoardResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	URL         types.String `tfsdk:"board_url"`
	Panels      types.List   `tfsdk:"panel"`
	Tags        types.Map    `tfsdk:"tags"`
}

type BoardPanelModel struct {
	PanelType  types.String `tfsdk:"type"`
	Position   types.List   `tfsdk:"position"`
	QueryPanel types.List   `tfsdk:"query_panel"`
	SLOPanel   types.List   `tfsdk:"slo_panel"`
}

var BoardPanelModelAttrType = map[string]attr.Type{
	"type": types.StringType,
	"position": types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: BoardPanelPositionModelAttrType,
		},
	},
	"query_panel": types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: QueryPanelModelAttrType,
		},
	},
	"slo_panel": types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: SLOPanelModelAttrType,
		},
	},
}

type BoardPanelPositionModel struct {
	XCoordinate types.Int64 `tfsdk:"x_coordinate"`
	YCoordinate types.Int64 `tfsdk:"y_coordinate"`
	Height      types.Int64 `tfsdk:"height"`
	Width       types.Int64 `tfsdk:"width"`
}

var BoardPanelPositionModelAttrType = map[string]attr.Type{
	"x_coordinate": types.Int64Type,
	"y_coordinate": types.Int64Type,
	"height":       types.Int64Type,
	"width":        types.Int64Type,
}

type QueryPanelModel struct {
	QueryID               types.String `tfsdk:"query_id"`
	QueryAnnotationID     types.String `tfsdk:"query_annotation_id"`
	QueryStyle            types.String `tfsdk:"query_style"`
	VisualizationSettings types.List   `tfsdk:"visualization_settings"`
}

var QueryPanelModelAttrType = map[string]attr.Type{
	"query_id":            types.StringType,
	"query_annotation_id": types.StringType,
	"query_style":         types.StringType,
	"visualization_settings": types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: VisualizationSettingsModelAttrType,
		},
	},
}

type SLOPanelModel struct {
	SLOID types.String `tfsdk:"slo_id"`
}

var SLOPanelModelAttrType = map[string]attr.Type{
	"slo_id": types.StringType,
}

type VisualizationSettingsModel struct {
	UseUTCXAxis          types.Bool `tfsdk:"use_utc_xaxis"`
	HideMarkers          types.Bool `tfsdk:"hide_markers"`
	HideHovers           types.Bool `tfsdk:"hide_hovers"`
	PreferOverlaidCharts types.Bool `tfsdk:"prefer_overlaid_charts"`
	HideCompare          types.Bool `tfsdk:"hide_compare"`
	Charts               types.List `tfsdk:"chart"` // List of ChartSettingsModel
}

var VisualizationSettingsModelAttrType = map[string]attr.Type{
	"use_utc_xaxis":          types.BoolType,
	"hide_markers":           types.BoolType,
	"hide_hovers":            types.BoolType,
	"prefer_overlaid_charts": types.BoolType,
	"hide_compare":           types.BoolType,
	"chart": types.ListType{
		ElemType: types.ObjectType{
			AttrTypes: ChartSettingsModelAttrType,
		},
	},
}

type ChartSettingsModel struct {
	ChartType         types.String `tfsdk:"chart_type"`
	ChartIndex        types.Int64  `tfsdk:"chart_index"`
	OmitMissingValues types.Bool   `tfsdk:"omit_missing_values"`
	LogScale          types.Bool   `tfsdk:"use_log_scale"`
}

var ChartSettingsModelAttrType = map[string]attr.Type{
	"chart_type":          types.StringType,
	"chart_index":         types.Int64Type,
	"omit_missing_values": types.BoolType,
	"use_log_scale":       types.BoolType,
}
