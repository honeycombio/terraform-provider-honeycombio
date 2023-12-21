package honeycombio

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func newWebhookRecipient() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWebhookRecipientCreate,
		ReadContext:   resourceWebhookRecipientRead,
		UpdateContext: resourceWebhookRecipientUpdate,
		DeleteContext: resourceWebhookRecipientDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Honeycomb Webhook Recipient allows you to define and manage a Webhook recipient that can be used by Triggers or BurnAlerts notifications.",

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Webhook Integration to create",
			},
			"secret": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The secret to include when sending the notification to the webhook",
			},
			"url": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The URL of the endpoint to send the integration to",
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			},
		},
	}
}

func resourceWebhookRecipientCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return createRecipient(ctx, d, meta, honeycombio.RecipientTypeWebhook)
}

func resourceWebhookRecipientRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return readRecipient(ctx, d, meta, honeycombio.RecipientTypeWebhook)
}

func resourceWebhookRecipientUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return updateRecipient(ctx, d, meta, honeycombio.RecipientTypeWebhook)
}

func resourceWebhookRecipientDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return deleteRecipient(ctx, d, meta)
}
