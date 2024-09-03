package honeycombio

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

var channelRegex = regexp.MustCompile(`^#.*|^@.*|^(C|D|G)[A-Z0-9]{6,}$`)

func newSlackRecipient() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSlackRecipientCreate,
		ReadContext:   resourceSlackRecipientRead,
		UpdateContext: resourceSlackRecipientUpdate,
		DeleteContext: resourceSlackRecipientDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Honeycomb Slack Recipient allows you to define and manage a Slack channel or user recipient that can be used by Triggers or BurnAlerts notifications.",

		Schema: map[string]*schema.Schema{
			"channel": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The Slack channel or username to send the notification to. Must begin with `#` or `@`.",
				ValidateFunc: validation.StringMatch(channelRegex, "channel must begin with `#` or `@` or be a valid channel id e.g. `CABC123DEF`"),
			},
		},
	}
}

func resourceSlackRecipientCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return createRecipient(ctx, d, meta, honeycombio.RecipientTypeSlack)
}

func resourceSlackRecipientRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return readRecipient(ctx, d, meta, honeycombio.RecipientTypeSlack)
}

func resourceSlackRecipientUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return updateRecipient(ctx, d, meta, honeycombio.RecipientTypeSlack)
}

func resourceSlackRecipientDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return deleteRecipient(ctx, d, meta)
}
