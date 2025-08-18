package honeycombio

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
)

func dataSourceHoneycombioSlackRecipient() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHoneycombioSlackRecipientRead,

		Schema: map[string]*schema.Schema{
			"dataset": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(helper.AsStringSlice(honeycombio.TriggerRecipientTypes()), false),
			},
			"target": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		DeprecationMessage: "Use honeycombio_recipient data source instead",
	}
}

func dataSourceHoneycombioSlackRecipientRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}
	dataset := d.Get("dataset").(string)

	triggers, err := client.Triggers.List(ctx, dataset)
	if err != nil {
		return diagFromErr(err)
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
