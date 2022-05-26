package honeycombio

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func dataSourceHoneycombioRecipient() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHoneycombioRecipientRead,

		Schema: map[string]*schema.Schema{
			"dataset": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(recipientTypeStrings("trigger"), false),
			},
			"target": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceHoneycombioRecipientRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)
	dataset := d.Get("dataset").(string)

	triggers, err := client.Triggers.List(ctx, dataset)
	if err != nil {
		return diag.FromErr(err)
	}

	searchType := honeycombio.RecipientType(d.Get("type").(string))
	searchTarget := d.Get("target").(string)

	for _, t := range triggers {
		for _, r := range t.Recipients {
			if r.Type == searchType && r.Target == searchTarget {
				d.SetId(r.ID)
				return nil
			}
		}
	}

	return diag.Errorf("could not find a trigger recipient in \"%s\" with type = \"%s\" and target = \"%s\"", dataset, searchType, searchTarget)
}
