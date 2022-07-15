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
	client := meta.(*honeycombio.Client)

	r, err := expandRecipient(honeycombio.RecipientTypeWebhook, d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, err = client.Recipients.Create(ctx, r)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.ID)
	return resourceWebhookRecipientRead(ctx, d, meta)
}

func resourceWebhookRecipientRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	r, err := client.Recipients.Get(ctx, d.Id())
	if err == honeycombio.ErrNotFound {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.ID)
	d.Set("name", r.Details.WebhookName)
	d.Set("secret", r.Details.WebhookSecret)
	d.Set("url", r.Details.WebhookURL)

	return nil
}

func resourceWebhookRecipientUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	r, err := expandRecipient(honeycombio.RecipientTypeWebhook, d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, err = client.Recipients.Update(ctx, r)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.ID)
	return resourceWebhookRecipientRead(ctx, d, meta)
}

func resourceWebhookRecipientDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	err := client.Recipients.Delete(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
