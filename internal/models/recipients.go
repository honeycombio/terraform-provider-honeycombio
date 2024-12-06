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
	Templates types.Set    `tfsdk:"template"` // WebhookTemplateModel
	Variables types.Set    `tfsdk:"variable"` // TemplateVariableModel
	Headers   types.Set    `tfsdk:"header"`   //WebhookHeaderModel
}

type WebhookTemplateModel struct {
	Type types.String `tfsdk:"type"`
	Body types.String `tfsdk:"body"`
}

var WebhookTemplateAttrType = map[string]attr.Type{
	"type": types.StringType,
	"body": types.StringType,
}

type TemplateVariableModel struct {
	Name         types.String `tfsdk:"name"`
	DefaultValue types.String `tfsdk:"default_value"`
}

var TemplateVariableAttrType = map[string]attr.Type{
	"name":          types.StringType,
	"default_value": types.StringType,
}

type WebhookHeaderModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

var WebhookHeaderAttrType = map[string]attr.Type{
	"name":  types.StringType,
	"value": types.StringType,
}
