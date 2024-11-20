package models

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WebhookRecipientModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Secret    types.String `tfsdk:"secret"`
	URL       types.String `tfsdk:"url"`
	Templates types.Set    `tfsdk:"templates"` // WebhookTemplateModel
}

type WebhookTemplateModel struct {
	Type types.String `tfsdk:"type"`
	Body types.String `tfsdk:"body"`
}

var WebhookTemplateAttrType = map[string]attr.Type{
	"type": types.StringType,
	"body": types.StringType,
}
