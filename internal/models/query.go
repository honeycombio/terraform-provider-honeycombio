package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type QueryResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Dataset   types.String `tfsdk:"dataset"`
	QueryJson types.String `tfsdk:"query_json"`
}
