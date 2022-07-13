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
		Description: "Honeycomb Email Recipient allows you to define and manage an email recipient that will can be used by Triggers or BurnAlerts notifications.",

		Schema: map[string]*schema.Schema{
			"address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The email address to send the notification to",
				ValidateFunc: func(v interface{}, key string) (warns []string, errs []error) {
					if _, err := mail.ParseAddress(v.(string)); err != nil {
						errs = append(errs, fmt.Errorf("unable to parse address \"%v\"", v))
					}
					return
				},
			},
		},
	}
}

func resourceEmailRecipientCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	r, err := expandRecipient(honeycombio.RecipientTypeEmail, d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, err = client.Recipients.Create(ctx, r)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.ID)
	return resourceEmailRecipientRead(ctx, d, meta)
}

func resourceEmailRecipientRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	r, err := client.Recipients.Get(ctx, d.Id())
	if err == honeycombio.ErrNotFound {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.ID)
	d.Set("address", r.Details.EmailAddress)

	return nil
}

func resourceEmailRecipientUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	r, err := expandRecipient(honeycombio.RecipientTypeEmail, d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, err = client.Recipients.Update(ctx, r)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.ID)
	return resourceEmailRecipientRead(ctx, d, meta)
}

func resourceEmailRecipientDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	err := client.Recipients.Delete(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
