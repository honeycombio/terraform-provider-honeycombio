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
	return createRecipient(ctx, d, meta, honeycombio.RecipientTypePagerDuty)
}

func resourcePDRecipientRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return readRecipient(ctx, d, meta, honeycombio.RecipientTypePagerDuty)
}

func resourcePDRecipientUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return updateRecipient(ctx, d, meta, honeycombio.RecipientTypePagerDuty)
}

func resourcePDRecipientDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return deleteRecipient(ctx, d, meta)
}
