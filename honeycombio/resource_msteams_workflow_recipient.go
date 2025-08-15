package honeycombio

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func newMSTeamsWorkflowRecipient() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMSTeamsWorkflowRecipientCreate,
		ReadContext:   resourceMSTeamsWorkflowRecipientRead,
		UpdateContext: resourceMSTeamsWorkflowRecipientUpdate,
		DeleteContext: resourceMSTeamsWorkflowRecipientDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Honeycomb MSTeams Workflow Recipient allows you to define and manage an MSTeams Workflows recipient that can be used by Triggers or BurnAlerts notifications.",

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the recipient.",
			},
			"url": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The Teams Workflow URL to send the notification to.",
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			},
		},
	}
}

func resourceMSTeamsWorkflowRecipientCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return createRecipient(ctx, d, meta, honeycombio.RecipientTypeMSTeamsWorkflow)
}

func resourceMSTeamsWorkflowRecipientRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return readRecipient(ctx, d, meta, honeycombio.RecipientTypeMSTeamsWorkflow)
}

func resourceMSTeamsWorkflowRecipientUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return updateRecipient(ctx, d, meta, honeycombio.RecipientTypeMSTeamsWorkflow)
}

func resourceMSTeamsWorkflowRecipientDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return deleteRecipient(ctx, d, meta)
}
