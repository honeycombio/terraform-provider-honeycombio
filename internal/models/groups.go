package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GroupResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

type GroupGrantResourceModel struct {
	ID        types.String `tfsdk:"id"`
	GroupID   types.String `tfsdk:"group_id"`
	DatasetID types.String `tfsdk:"dataset_id"`
}

type GroupMembersResourceModel struct {
	ID      types.String `tfsdk:"id"`
	GroupID types.String `tfsdk:"group_id"`
	UserIDs types.Set    `tfsdk:"user_ids"` // types.String
}
