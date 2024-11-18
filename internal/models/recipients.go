package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WebhookRecipientModel struct {
	ID              types.String         `tfsdk:"id"`
	Name            types.String         `tfsdk:"name"`
	Secret          types.String         `tfsdk:"secret"`
	URL             types.String         `tfsdk:"url"`
	WebhookPayloads WebhookPayloadsModel `tfsdk:"webhook_payloads"`
}
