package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/filter"
)

type SLOResourceModel struct {
	ID               types.String  `tfsdk:"id"`
	Name             types.String  `tfsdk:"name"`
	Description      types.String  `tfsdk:"description"`
	Dataset          types.String  `tfsdk:"dataset"`
	Datasets         types.Set     `tfsdk:"datasets"`
	SLI              types.String  `tfsdk:"sli"`
	TargetPercentage types.Float64 `tfsdk:"target_percentage"`
	TimePeriod       types.Int64   `tfsdk:"time_period"`
	Tags             types.Map     `tfsdk:"tags"`
}

type SLOsDataSourceModel struct {
	ID           types.String               `tfsdk:"id"`
	Dataset      types.String               `tfsdk:"dataset"`
	DetailFilter []filter.DetailFilterModel `tfsdk:"detail_filter"`
	IDs          []types.String             `tfsdk:"ids"`
}

type SLODataSourceModel struct {
	ID               types.String   `tfsdk:"id"`
	Dataset          types.String   `tfsdk:"dataset"`
	Datasets         []types.String `tfsdk:"datasets"`
	Name             types.String   `tfsdk:"name"`
	Description      types.String   `tfsdk:"description"`
	SLI              types.String   `tfsdk:"sli"`
	Tags             types.Map      `tfsdk:"tags"`
	TargetPercentage types.Float64  `tfsdk:"target_percentage"`
	TimePeriod       types.Int64    `tfsdk:"time_period"`
}
