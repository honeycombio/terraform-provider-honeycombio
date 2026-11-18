package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GroupResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}
