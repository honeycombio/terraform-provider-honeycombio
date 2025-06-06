package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/filter"
)

type DatasetResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Slug            types.String `tfsdk:"slug"`
	Description     types.String `tfsdk:"description"`
	ExpandJSONDepth types.Int32  `tfsdk:"expand_json_depth"`
	DeleteProtected types.Bool   `tfsdk:"delete_protected"`
	CreatedAt       types.String `tfsdk:"created_at"`
	LastWrittenAt   types.String `tfsdk:"last_written_at"`
}

type DatasetsDataSourceModel struct {
	ID           types.String               `tfsdk:"id"`
	StartsWith   types.String               `tfsdk:"starts_with"`
	DetailFilter []filter.DetailFilterModel `tfsdk:"detail_filter"`
	Names        []types.String             `tfsdk:"names"`
	Slugs        []types.String             `tfsdk:"slugs"`
}
