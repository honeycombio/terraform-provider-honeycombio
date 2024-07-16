package honeycombio

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
)

func dataSourceHoneycombioRecipient() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHoneycombioRecipientRead,

		Description: `
'honeycombio_recipient' data source provides details about a specific recipient.

The ID of an existing recipient can be used when adding recipients to triggers or burn alerts.

Note: Terraform will fail unless exactly one recipient is returned by the search. Ensure
that your search is specific enough to return a single recipient ID only.
If you want to match multiple recipients, use the 'honeycombio_recipients' data source instead.
`,

		Schema: map[string]*schema.Schema{
			"dataset": {
				Type:        schema.TypeString,
				Optional:    true,
				Deprecated:  "Recipients are now a Team-level construct. The provided 'dataset' value is being ignored and should be removed.",
				Description: "Deprecated: recipients are now a Team-level construct. Any provided 'dataset' value will be ignored.",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The type of recipient.",
				ValidateFunc: validation.StringInSlice(helper.AsStringSlice(honeycombio.RecipientTypes()), false),
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
			// type-specific generated attributes
			"address": {
				Type:        schema.TypeString,
				Computed:    true,
				Required:    false,
				Optional:    false,
				Description: "The email recipient's address -- if of type `email`",
			},
			"channel": {
				Type:        schema.TypeString,
				Computed:    true,
				Required:    false,
				Optional:    false,
				Description: "The Slack recipient's channel -- if of type `slack`",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Required:    false,
				Optional:    false,
				Description: "The webhook recipient's name -- if of type `webhook` or `msteams`",
			},
			"secret": {
				Type:        schema.TypeString,
				Computed:    true,
				Required:    false,
				Optional:    false,
				Sensitive:   true,
				Description: "The webhook recipient's secret -- if of type `webhook`",
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Required:    false,
				Optional:    false,
				Description: "The webhook recipient's URL -- if of type `webhook` or `msteams`",
			},
			"integration_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Required:    false,
				Optional:    false,
				Sensitive:   true,
				Description: "The PagerDuty recipient's key -- if of type `pagerduty`",
			},
			"integration_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Required:    false,
				Optional:    false,
				Description: "The PagerDuty recipient's name -- if of type `pagerduty`",
			},
		},
	}
}

func dataSourceHoneycombioRecipientRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	rcpts, err := client.Recipients.List(ctx)
	if err != nil {
		return diagFromErr(err)
	}
	matchType := honeycombio.RecipientType(d.Get("type").(string))

	rcptFilter := &recipientFilter{Type: matchType}
	if v, ok := d.GetOk("target"); ok {
		// deprecated argument to be removed in future
		rcptFilter = &recipientFilter{Value: honeycombio.ToPtr(v.(string))}
	}
	if v, ok := d.GetOk("detail_filter"); ok {
		rcptFilter = expandRecipientFilter(v.([]interface{}))
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
		return diag.Errorf("your recipient query returned more than one result. Please try a more specific search criteria.")
	}
	rcpt := filteredRcpts[0]
	d.SetId(rcpt.ID)
	// type-specific generated attributes
	switch matchType {
	case honeycombio.RecipientTypeEmail:
		d.Set("address", rcpt.Details.EmailAddress)
	case honeycombio.RecipientTypeSlack:
		d.Set("channel", rcpt.Details.SlackChannel)
	case honeycombio.RecipientTypeMSTeams:
		d.Set("name", rcpt.Details.WebhookName)
		d.Set("url", rcpt.Details.WebhookURL)
	case honeycombio.RecipientTypeWebhook:
		d.Set("name", rcpt.Details.WebhookName)
		d.Set("secret", rcpt.Details.WebhookSecret)
		d.Set("url", rcpt.Details.WebhookURL)
	case honeycombio.RecipientTypePagerDuty:
		d.Set("integration_key", rcpt.Details.PDIntegrationKey)
		d.Set("integration_name", rcpt.Details.PDIntegrationName)
	}

	return nil
}
