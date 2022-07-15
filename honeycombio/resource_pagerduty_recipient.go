package honeycombio

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func newPDRecipient() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePDRecipientCreate,
		ReadContext:   resourcePDRecipientRead,
		UpdateContext: resourcePDRecipientUpdate,
		DeleteContext: resourcePDRecipientDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Honeycomb PagerDuty Recipient allows you to define and manage a PagerDuty recipient that can be used by Triggers or BurnAlerts notifications.",

		Schema: map[string]*schema.Schema{
			"integration_key": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The key of the PagerDuty Integration to send the notification to",
				ValidateFunc: validation.StringLenBetween(32, 32),
			},
			"integration_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the PagerDuty Integration to send the notification to",
			},
		},
	}
}

func resourcePDRecipientCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	r, err := expandRecipient(honeycombio.RecipientTypePagerDuty, d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, err = client.Recipients.Create(ctx, r)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.ID)
	return resourcePDRecipientRead(ctx, d, meta)
}

func resourcePDRecipientRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	r, err := client.Recipients.Get(ctx, d.Id())
	if err == honeycombio.ErrNotFound {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.ID)
	d.Set("integration_key", r.Details.PDIntegrationKey)
	d.Set("integration_name", r.Details.PDIntegrationName)

	return nil
}

func resourcePDRecipientUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	r, err := expandRecipient(honeycombio.RecipientTypePagerDuty, d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, err = client.Recipients.Update(ctx, r)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.ID)
	return resourcePDRecipientRead(ctx, d, meta)
}

func resourcePDRecipientDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	err := client.Recipients.Delete(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
