package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type WebhookPayloadsModel struct {
	PayloadTemplates PayloadTemplatesModel `tfsdk:"payload_templates"`
}

type PayloadTemplatesModel struct {
	Trigger        *PayloadTemplateModel `tfsdk:"trigger,omitempty"`
	ExhaustionTime *PayloadTemplateModel `tfsdk:"exhaustion_time,omitempty"`
	BudgetRate     *PayloadTemplateModel `tfsdk:"budget_rate,omitempty"`
}

type PayloadTemplateModel struct {
	Body types.String `tfsdk:"body"`
}
