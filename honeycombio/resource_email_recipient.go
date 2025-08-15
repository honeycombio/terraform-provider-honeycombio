package honeycombio

import (
	"context"
	"fmt"
	"net/mail"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func newEmailRecipient() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEmailRecipientCreate,
		ReadContext:   resourceEmailRecipientRead,
		UpdateContext: resourceEmailRecipientUpdate,
		DeleteContext: resourceEmailRecipientDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Honeycomb Email Recipient allows you to define and manage an Email recipient that can be used by Triggers or BurnAlerts notifications.",

		Schema: map[string]*schema.Schema{
			"address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The email address to send the notification to",
				ValidateFunc: func(v any, key string) (warns []string, errs []error) {
					if _, err := mail.ParseAddress(v.(string)); err != nil {
						errs = append(errs, fmt.Errorf("unable to parse address \"%v\"", v))
					}
					return
				},
			},
		},
	}
}

func resourceEmailRecipientCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return createRecipient(ctx, d, meta, honeycombio.RecipientTypeEmail)
}

func resourceEmailRecipientRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return readRecipient(ctx, d, meta, honeycombio.RecipientTypeEmail)
}

func resourceEmailRecipientUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return updateRecipient(ctx, d, meta, honeycombio.RecipientTypeEmail)
}

func resourceEmailRecipientDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return deleteRecipient(ctx, d, meta)
}
