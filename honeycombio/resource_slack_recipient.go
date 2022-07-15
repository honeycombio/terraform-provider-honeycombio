package honeycombio

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

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
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^(#|@).+`), "channel must begin with `#` or `@`"),
			},
		},
	}
}

func resourceSlackRecipientCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	r, err := expandRecipient(honeycombio.RecipientTypeSlack, d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, err = client.Recipients.Create(ctx, r)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.ID)
	return resourceSlackRecipientRead(ctx, d, meta)
}

func resourceSlackRecipientRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	r, err := client.Recipients.Get(ctx, d.Id())
	if err == honeycombio.ErrNotFound {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.ID)
	d.Set("channel", r.Details.SlackChannel)

	return nil
}

func resourceSlackRecipientUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	r, err := expandRecipient(honeycombio.RecipientTypeSlack, d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, err = client.Recipients.Update(ctx, r)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.ID)
	return resourceSlackRecipientRead(ctx, d, meta)
}

func resourceSlackRecipientDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	err := client.Recipients.Delete(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
