package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type UserDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Email types.String `tfsdk:"email"`
	Role  types.String `tfsdk:"role"`
}

type UsersDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	GroupID types.String `tfsdk:"group_id"`
	Users   []UserModel  `tfsdk:"users"`
}

// UserModel is one user in the users list of honeycombio_users.
type UserModel struct {
	ID    types.String `tfsdk:"id"`
	Email types.String `tfsdk:"email"`
	Role  types.String `tfsdk:"role"`
}
