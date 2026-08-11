package models

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type APIKeyResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Type             types.String `tfsdk:"type"`
	EnvironmentID    types.String `tfsdk:"environment_id"`
	VisibleToMembers types.Bool   `tfsdk:"visible_to_members"`
	Disabled         types.Bool   `tfsdk:"disabled"`
	Permissions      types.List   `tfsdk:"permissions"` // APIKeyPermissionModel
	Secret           types.String `tfsdk:"secret"`
	Key              types.String `tfsdk:"key"`
}

type APIKeyPermissionModel struct {
	SendEvents          types.Bool `tfsdk:"send_events"`
	CreateDatasets      types.Bool `tfsdk:"create_datasets"`
	ManageQueries       types.Bool `tfsdk:"manage_queries"`
	RunQueries          types.Bool `tfsdk:"run_queries"`
	ReadServiceMaps     types.Bool `tfsdk:"read_service_maps"`
	ManagePublicBoards  types.Bool `tfsdk:"manage_public_boards"`
	ManagePrivateBoards types.Bool `tfsdk:"manage_private_boards"`
	ManageSLOs          types.Bool `tfsdk:"manage_slos"`
	ManageTriggers      types.Bool `tfsdk:"manage_triggers"`
	ManageRecipients    types.Bool `tfsdk:"manage_recipients"`
	ManageMarkers       types.Bool `tfsdk:"manage_markers"`
}

var APIKeyPermissionsAttrType = map[string]attr.Type{
	"send_events":           types.BoolType,
	"create_datasets":       types.BoolType,
	"manage_queries":        types.BoolType,
	"run_queries":           types.BoolType,
	"read_service_maps":     types.BoolType,
	"manage_public_boards":  types.BoolType,
	"manage_private_boards": types.BoolType,
	"manage_slos":           types.BoolType,
	"manage_triggers":       types.BoolType,
	"manage_recipients":     types.BoolType,
	"manage_markers":        types.BoolType,
}
