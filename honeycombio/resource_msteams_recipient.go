package honeycombio

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func newMSTeamsRecipient() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMSTeamsRecipientCreate,
		ReadContext:   resourceMSTeamsRecipientRead,
		UpdateContext: resourceMSTeamsRecipientUpdate,
		DeleteContext: resourceMSTeamsRecipientDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Honeycomb MSTeams Recipient allows you to define and manage an MSTeams recipient that can be used by Triggers or BurnAlerts notifications.",

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the MSTeams Integration to create",
			},
			"url": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The Incoming Webhook URL to send the notification to",
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			},
		},
	}
}

func resourceMSTeamsRecipientCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return createRecipient(ctx, d, meta, honeycombio.RecipientTypeMSTeams)
}

func resourceMSTeamsRecipientRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return readRecipient(ctx, d, meta, honeycombio.RecipientTypeMSTeams)
}

func resourceMSTeamsRecipientUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return updateRecipient(ctx, d, meta, honeycombio.RecipientTypeMSTeams)
}

func resourceMSTeamsRecipientDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return deleteRecipient(ctx, d, meta)
}
