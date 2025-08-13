package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ColumnResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Dataset       types.String `tfsdk:"dataset"`
	Hidden        types.Bool   `tfsdk:"hidden"`
	Description   types.String `tfsdk:"description"`
	Type          types.String `tfsdk:"type"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
	LastWrittenAt types.String `tfsdk:"last_written_at"`
}
