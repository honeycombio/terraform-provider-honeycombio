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

func dataSourceHoneycombioSlackRecipientRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}
	dataset, ok := d.Get("dataset").(string)
	if !ok {
		return diag.Errorf("dataset must be a string")
	}

	triggers, err := client.Triggers.List(ctx, dataset)
	if err != nil {
		return diagFromErr(err)
	}

	typeStr, ok := d.Get("type").(string)
	if !ok {
		return diag.Errorf("type must be a string")
	}
	searchType := honeycombio.RecipientType(typeStr)

	targetStr, ok := d.Get("target").(string)
	if !ok {
		return diag.Errorf("target must be a string")
	}
	searchTarget := targetStr

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
