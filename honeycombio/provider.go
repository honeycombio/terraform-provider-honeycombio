package honeycombio

import (
	"context"
	"errors"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/log"
)

func init() {
	// set global description kind to markdown, as described in:
	// https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html#support-for-resource-level-and-field-level-descriptions
	schema.DescriptionKind = schema.StringMarkdown
}

func Provider(version string) *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Description: "The Honeycomb API key to use. It can also be set via the `HONEYCOMB_API_KEY` or `HONEYCOMBIO_APIKEY` environment variables.",
				Optional:    true,
				Sensitive:   true,
			},
			"api_key_id": { // unused in the provider but required to be set for the MuxServer
				Type:        schema.TypeString,
				Description: "The ID portion of the Honeycomb Management API key to use. It can also be set via the `HONEYCOMB_KEY_ID` environment variable.",
				Optional:    true,
				Sensitive:   false,
			},
			"api_key_secret": { // unused in the provider but required to be set for the MuxServer
				Type:        schema.TypeString,
				Description: "The secret portion of the Honeycomb Management API key to use. It can also be set via the `HONEYCOMB_KEY_SECRET` environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"api_url": {
				Type:        schema.TypeString,
				Description: "Override the URL of the Honeycomb API. Defaults to `https://api.honeycomb.io`. It can also be set via the `HONEYCOMB_API_ENDPOINT` environment variable.",
				Optional:    true,
			},
			"debug": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Enable the API client's debug logs. By default, a `TF_LOG` setting of debug or higher will enable this.",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"honeycombio_datasets":          dataSourceHoneycombioDatasets(),
			"honeycombio_column":            dataSourceHoneycombioColumn(),
			"honeycombio_columns":           dataSourceHoneycombioColumns(),
			"honeycombio_query_result":      dataSourceHoneycombioQueryResult(),
			"honeycombio_trigger_recipient": dataSourceHoneycombioSlackRecipient(),
			"honeycombio_recipient":         dataSourceHoneycombioRecipient(),
			"honeycombio_recipients":        dataSourceHoneycombioRecipients(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"honeycombio_board":               newBoard(),
			"honeycombio_column":              newColumn(),
			"honeycombio_dataset":             newDataset(),
			"honeycombio_dataset_definition":  newDatasetDefinition(),
			"honeycombio_derived_column":      newDerivedColumn(),
			"honeycombio_marker":              newMarker(),
			"honeycombio_marker_setting":      newMarkerSetting(),
			"honeycombio_query_annotation":    newQueryAnnotation(),
			"honeycombio_email_recipient":     newEmailRecipient(),
			"honeycombio_pagerduty_recipient": newPDRecipient(),
			"honeycombio_msteams_recipient":   newMSTeamsRecipient(),
			"honeycombio_slack_recipient":     newSlackRecipient(),
			"honeycombio_webhook_recipient":   newWebhookRecipient(),
			"honeycombio_slo":                 newSLO(),
		},
	}

	provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		apiKey := os.Getenv(honeycombio.DefaultAPIKeyEnv)
		if apiKey == "" {
			// fall through to legacy env var
			//nolint:staticcheck
			apiKey = os.Getenv(honeycombio.LegacyAPIKeyEnv)
		}
		if v, ok := d.GetOk("api_key"); ok {
			apiKey = v.(string)
		}
		debug := log.IsDebugOrHigher()
		if v, ok := d.GetOk("debug"); ok {
			debug = v.(bool)
		}

		// if the API key is set, use it to create the client
		// we now rely on the Framework version of the provider to validate the configuration
		if apiKey != "" {
			config := &honeycombio.Config{
				APIKey:    apiKey,
				APIUrl:    d.Get("api_url").(string),
				UserAgent: provider.UserAgent("terraform-provider-honeycombio", version),
				Debug:     debug,
			}
			c, err := honeycombio.NewClientWithConfig(config)
			if err != nil {
				return nil, diag.FromErr(err)
			}
			return c, nil
		}

		return nil, nil
	}

	return provider
}

func getConfiguredClient(meta any) (*honeycombio.Client, error) {
	client, ok := meta.(*honeycombio.Client)
	if !ok || client == nil {
		return nil, errors.New("No v1 API client configured for this provider. " +
			"Set the `api_key` attribute in the provider's configuration, " +
			"or set the HONEYCOMB_API_KEY environment variable.")
	}
	return client, nil
}
