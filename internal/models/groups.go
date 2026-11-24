package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// GroupModel is the model of the honeycombio_group resource and data source.
type GroupModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

type GroupGrantResourceModel struct {
	ID            types.String `tfsdk:"id"`
	GroupID       types.String `tfsdk:"group_id"`
	DatasetID     types.String `tfsdk:"dataset_id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
}

// GroupGrantsModel is the model of the honeycombio_group_grants resource
// and data source.
type GroupGrantsModel struct {
	ID         types.String `tfsdk:"id"`
	GroupID    types.String `tfsdk:"group_id"`
	DatasetIDs types.Set    `tfsdk:"dataset_ids"` // types.String
}

type GroupMembersResourceModel struct {
	ID      types.String `tfsdk:"id"`
	GroupID types.String `tfsdk:"group_id"`
	UserIDs types.Set    `tfsdk:"user_ids"` // types.String
}
