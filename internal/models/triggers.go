package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type TriggerResourceModel struct {
	ID                 types.String                     `tfsdk:"id"`
	Name               types.String                     `tfsdk:"name"`
	Dataset            types.String                     `tfsdk:"dataset"`
	Description        types.String                     `tfsdk:"description"`
	Disabled           types.Bool                       `tfsdk:"disabled"`
	QueryID            types.String                     `tfsdk:"query_id"`
	AlertType          types.String                     `tfsdk:"alert_type"`
	Frequency          types.Int64                      `tfsdk:"frequency"`
	Threshold          []TriggerThresholdModel          `tfsdk:"threshold"`
	Recipients         []NotificationRecipientModel     `tfsdk:"recipient"`
	EvaluationSchedule []TriggerEvaluationScheduleModel `tfsdk:"evaluation_schedule"`
}

type TriggerThresholdModel struct {
	Op            types.String  `tfsdk:"op"`
	Value         types.Float64 `tfsdk:"value"`
	ExceededLimit types.Int64   `tfsdk:"exceeded_limit"`
}

type TriggerEvaluationScheduleModel struct {
	DaysOfWeek []types.String `tfsdk:"days_of_week"`
	StartTime  types.String   `tfsdk:"start_time"`
	EndTime    types.String   `tfsdk:"end_time"`
}
