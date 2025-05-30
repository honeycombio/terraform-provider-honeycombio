package models

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TriggerResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Dataset            types.String `tfsdk:"dataset"`
	Description        types.String `tfsdk:"description"`
	Disabled           types.Bool   `tfsdk:"disabled"`
	QueryID            types.String `tfsdk:"query_id"`
	QueryJson          types.String `tfsdk:"query_json"`
	AlertType          types.String `tfsdk:"alert_type"`
	Frequency          types.Int64  `tfsdk:"frequency"`
	Threshold          types.List   `tfsdk:"threshold"`           // TriggerThresholdModel
	Recipients         types.Set    `tfsdk:"recipient"`           // NotificationRecipientModel
	EvaluationSchedule types.List   `tfsdk:"evaluation_schedule"` // TriggerEvaluationScheduleModel
	BaselineDetails    types.List   `tfsdk:"baseline_details"`
	Tags               types.Map    `tfsdk:"tags"`
}

type TriggerBaselineDetailsModel struct {
	Type          types.String `tfsdk:"type"`
	OffsetMinutes types.Int64  `tfsdk:"offset_minutes"`
}

var TriggerBaselineDetailsAttrType = map[string]attr.Type{
	"type":           types.StringType,
	"offset_minutes": types.Int64Type,
}

type TriggerThresholdModel struct {
	Op            types.String  `tfsdk:"op"`
	Value         types.Float64 `tfsdk:"value"`
	ExceededLimit types.Int64   `tfsdk:"exceeded_limit"`
}

var TriggerThresholdAttrType = map[string]attr.Type{
	"op":             types.StringType,
	"value":          types.Float64Type,
	"exceeded_limit": types.Int64Type,
}

type TriggerEvaluationScheduleModel struct {
	DaysOfWeek []types.String `tfsdk:"days_of_week"`
	StartTime  types.String   `tfsdk:"start_time"`
	EndTime    types.String   `tfsdk:"end_time"`
}

var TriggerEvaluationScheduleAttrType = map[string]attr.Type{
	"days_of_week": types.ListType{ElemType: types.StringType},
	"start_time":   types.StringType,
	"end_time":     types.StringType,
}
