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

		Description: `
'honeycombio_recipient' data source provides details about a specific recipient.

The ID of an existing recipient can be used when adding recipients to triggers or burn alerts.

Note: If more or less than a single match is returned by the search, Terraform will fail. Ensure
that your search is specific enough to return a single recipient ID only.
If you want to match multiple recipients, use the 'honeycombio_recipients' data source instead.
`,

		Schema: map[string]*schema.Schema{
			"dataset": {
				Type:        schema.TypeString,
				Optional:    true,
				Deprecated:  "Recpients are now a Team-level construct. The provided 'dataset' value is being ignored and should be removed.",
				Description: "Deprecated: recpients are now a Team-level construct. Any provided 'dataset' value will being ignored.",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The type of recipient.",
				ValidateFunc: validation.StringInSlice([]string{"email", "pagerduty", "slack", "webhook"}, false),
			},
			"target": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"detail_filter"},
				Description:   "Deprecated: use 'detail_filter' instead. Target of the trigger or burn alert, this has another meaning depending on the type of recipient.",
				Deprecated:    "Use of 'target' has been replaced by 'detail_filter'.",
			},
			"detail_filter": {
				Type:          schema.TypeList,
				Optional:      true,
				MinItems:      1,
				MaxItems:      1,
				Description:   "Attributes to filter the recipients with.",
				ConflictsWith: []string{"target"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The name of the detail field to filter by",
							ValidateFunc: validation.StringInSlice([]string{"address", "channel", "name", "integration_name", "url"}, false),
						},
						"value": {
							Type:          schema.TypeString,
							Optional:      true,
							Description:   "The value of the detail field to match on.",
							ConflictsWith: []string{"detail_filter.0.value_regex"},
							ValidateFunc:  validation.NoZeroValues,
						},
						"value_regex": {
							Type:          schema.TypeString,
							Optional:      true,
							Description:   "A regular expression string to apply to the value of the detail field to match on.",
							ConflictsWith: []string{"detail_filter.0.value"},
							ValidateFunc:  validation.StringIsValidRegExp,
						},
					},
				},
			},
		},
	}
}

func dataSourceHoneycombioRecipientRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	rcpts, err := client.Recipients.List(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	matchType := honeycombio.RecipientType(d.Get("type").(string))

	var rcptFilter *recipientFilter
	if v, ok := d.GetOk("target"); ok {
		// deprecated attribute to be removed in future
		target := v.(string)
		rcptFilter = &recipientFilter{Type: matchType, Value: &target}
	}
	if v, ok := d.GetOk("detail_filter"); ok {
		rcptFilter = expandRecipientFilter(v.([]interface{}))
		if rcptFilter.Type != matchType {
			return diag.Errorf("provided type doesn't match filter type")
		}
	}

	var filteredRcpts []honeycombio.Recipient
	for _, r := range rcpts {
		if rcptFilter.IsMatch(r) {
			filteredRcpts = append(filteredRcpts, r)
		}
	}

	if len(filteredRcpts) < 1 {
		return diag.Errorf("your recipient query returned no results.")
	}
	if len(filteredRcpts) > 1 {
		return diag.Errorf("your recipient query returned more than one result. Please try a more specific search critera.")
	}
	rcpt := filteredRcpts[0]
	d.SetId(rcpt.ID)

	return nil
}
