package models

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type APIKeyResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Type          types.String `tfsdk:"type"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	Disabled      types.Bool   `tfsdk:"disabled"`
	Permissions   types.List   `tfsdk:"permissions"` // APIKeyPermissionModel
	Secret        types.String `tfsdk:"secret"`
}

type APIKeyPermissionModel struct {
	CreateDatasets types.Bool `tfsdk:"create_datasets"`
}

var APIKeyPermissionsAttrType = map[string]attr.Type{
	"create_datasets": types.BoolType,
}
