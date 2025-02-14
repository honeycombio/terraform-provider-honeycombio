package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SLOsDataSourceModel struct {
	ID           types.String        `tfsdk:"id"`
	Dataset      types.String        `tfsdk:"dataset"`
	DetailFilter []DetailFilterModel `tfsdk:"detail_filter"`
	IDs          []types.String      `tfsdk:"ids"`
}

type SLODataSourceModel struct {
	ID               types.String   `tfsdk:"id"`
	Dataset          types.String   `tfsdk:"dataset"`
	Datasets         []types.String `tfsdk:"datasets"`
	Name             types.String   `tfsdk:"name"`
	Description      types.String   `tfsdk:"description"`
	SLI              types.String   `tfsdk:"sli"`
	TargetPercentage types.Float64  `tfsdk:"target_percentage"`
	TimePeriod       types.Int64    `tfsdk:"time_period"`
}
