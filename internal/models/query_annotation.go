package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type QueryAnnotationResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Dataset     types.String `tfsdk:"dataset"`
	QueryID     types.String `tfsdk:"query_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}