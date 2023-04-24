package honeycombio

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func init() {
	// set global description kind to markdown, as described in:
	// https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html#support-for-resource-level-and-field-level-descriptions
	schema.DescriptionKind = schema.StringMarkdown
}

// providerVersion represents the current version of the provider. It should be
// overwritten during the release process.
var providerVersion = "dev"

func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"HONEYCOMB_API_KEY", "HONEYCOMBIO_APIKEY"}, nil),
				Sensitive:   true,
			},
			"api_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"debug": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Enable the API client's debug logs. By default, a `TF_LOG` setting of debug or higher will enable this.",
				DefaultFunc: func() (interface{}, error) {
					// use provider environment's configured log level
					return logging.IsDebugOrHigher(), nil
				},
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"honeycombio_datasets":            dataSourceHoneycombioDatasets(),
			"honeycombio_column":              dataSourceHoneycombioColumn(),
			"honeycombio_columns":             dataSourceHoneycombioColumns(),
			"honeycombio_query_result":        dataSourceHoneycombioQueryResult(),
			"honeycombio_query_specification": dataSourceHoneycombioQuerySpec(),
			"honeycombio_trigger_recipient":   dataSourceHoneycombioSlackRecipient(),
			"honeycombio_recipient":           dataSourceHoneycombioRecipient(),
			"honeycombio_recipients":          dataSourceHoneycombioRecipients(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"honeycombio_board":               newBoard(),
			"honeycombio_burn_alert":          newBurnAlert(),
			"honeycombio_column":              newColumn(),
			"honeycombio_dataset":             newDataset(),
			"honeycombio_dataset_definition":  newDatasetDefinition(),
			"honeycombio_derived_column":      newDerivedColumn(),
			"honeycombio_marker":              newMarker(),
			"honeycombio_marker_setting":      newMarkerSetting(),
			"honeycombio_query":               newQuery(),
			"honeycombio_query_annotation":    newQueryAnnotation(),
			"honeycombio_email_recipient":     newEmailRecipient(),
			"honeycombio_pagerduty_recipient": newPDRecipient(),
			"honeycombio_slack_recipient":     newSlackRecipient(),
			"honeycombio_webhook_recipient":   newWebhookRecipient(),
			"honeycombio_slo":                 newSLO(),
			"honeycombio_trigger":             newTrigger(),
		},
	}

	provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		config := &honeycombio.Config{
			APIKey:    d.Get("api_key").(string),
			APIUrl:    d.Get("api_url").(string),
			UserAgent: provider.UserAgent("terraform-provider-honeycombio", providerVersion),
			Debug:     d.Get("debug").(bool),
		}
		c, err := honeycombio.NewClient(config)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return c, nil
	}

	return provider
}
