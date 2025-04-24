package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type QueryResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Dataset   types.String `tfsdk:"dataset"`
	QueryJson types.String `tfsdk:"query_json"`
}

type QuerySpecificationModel struct {
	ID                types.String                             `tfsdk:"id"`
	FilterCombination types.String                             `tfsdk:"filter_combination"`
	Breakdowns        []types.String                           `tfsdk:"breakdowns"`
	Limit             types.Int64                              `tfsdk:"limit"`
	TimeRange         types.Int64                              `tfsdk:"time_range"`
	StartTime         types.Int64                              `tfsdk:"start_time"`
	EndTime           types.Int64                              `tfsdk:"end_time"`
	Granularity       types.Int64                              `tfsdk:"granularity"`
	Calculations      []QuerySpecificationCalculationModel     `tfsdk:"calculation"`
	CalculatedFields  []QuerySpecificationCalculatedFieldModel `tfsdk:"calculated_field"`
	Filters           []QuerySpecificationFilterModel          `tfsdk:"filter"`
	Havings           []QuerySpecificationHavingModel          `tfsdk:"having"`
	Orders            []QuerySpecificationOrderModel           `tfsdk:"order"`
	Json              types.String                             `tfsdk:"json"` // Computed JSON query specification output
}

type QuerySpecificationCalculationModel struct {
	Column types.String `tfsdk:"column"`
	Op     types.String `tfsdk:"op"`
}

type QuerySpecificationCalculatedFieldModel struct {
	Name       types.String `tfsdk:"name"`
	Expression types.String `tfsdk:"expression"`
}

type QuerySpecificationFilterModel struct {
	Column types.String `tfsdk:"column"`
	Op     types.String `tfsdk:"op"`
	Value  types.String `tfsdk:"value"` // TODO: convert to types.DynamicType
}

type QuerySpecificationHavingModel struct {
	CalculateOp types.String  `tfsdk:"calculate_op"`
	Column      types.String  `tfsdk:"column"`
	Op          types.String  `tfsdk:"op"`
	Value       types.Float64 `tfsdk:"value"`
}

type QuerySpecificationOrderModel struct {
	Column types.String `tfsdk:"column"`
	Op     types.String `tfsdk:"op"`
	Order  types.String `tfsdk:"order"`
}
