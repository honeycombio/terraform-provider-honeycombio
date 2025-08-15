package honeycombio

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

// Deprecated: MSTeams Recipient is deprecated, and does not allow creation of new recipients.
func newMSTeamsRecipient() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMSTeamsRecipientCreate,
		ReadContext:   resourceMSTeamsRecipientRead,
		UpdateContext: resourceMSTeamsRecipientUpdate,
		DeleteContext: resourceMSTeamsRecipientDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		DeprecationMessage: "MSTeams Recipient is deprecated. " +
			"Creating new MSTeams recipients is no longer possible." +
			"Please use the `honeycombio_msteams_workflow_recipient` resource instead. ",
		Description: "Honeycomb MSTeams Recipient allows you to define and manage an MSTeams recipient that can be used by Triggers or BurnAlerts notifications.",

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the recipient.",
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

func resourceMSTeamsRecipientCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return diag.Diagnostics{diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Creating new MSTeams recipients is no longer possible.",
		Detail: "Microsoft has deprecated the Incoming Webhook feature, and as a result, " +
			" we are no longer able to create new MSTeams recipients. " +
			"Use the `honeycombio_msteams_workflow_recipient` resource instead.",
	}}
}

func resourceMSTeamsRecipientRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	//nolint:staticcheck
	return readRecipient(ctx, d, meta, honeycombio.RecipientTypeMSTeams)
}

func resourceMSTeamsRecipientUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	//nolint:staticcheck
	return updateRecipient(ctx, d, meta, honeycombio.RecipientTypeMSTeams)
}

func resourceMSTeamsRecipientDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return deleteRecipient(ctx, d, meta)
}
