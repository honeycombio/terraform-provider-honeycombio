package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/filter"
)

type EnvironmentResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Slug            types.String `tfsdk:"slug"`
	Description     types.String `tfsdk:"description"`
	Color           types.String `tfsdk:"color"`
	DeleteProtected types.Bool   `tfsdk:"delete_protected"`
}

type EnvironmentDataSourceModel struct {
	ID              types.String               `tfsdk:"id"`
	DetailFilter    []filter.DetailFilterModel `tfsdk:"detail_filter"`
	Name            types.String               `tfsdk:"name"`
	Slug            types.String               `tfsdk:"slug"`
	Description     types.String               `tfsdk:"description"`
	Color           types.String               `tfsdk:"color"`
	DeleteProtected types.Bool                 `tfsdk:"delete_protected"`
}

type EnvironmentsDataSourceModel struct {
	ID           types.String               `tfsdk:"id"`
	DetailFilter []filter.DetailFilterModel `tfsdk:"detail_filter"`
	IDs          []types.String             `tfsdk:"ids"`
}
