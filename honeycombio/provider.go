package honeycombio

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	honeycombio "github.com/kvrhdn/go-honeycombio"
)

// providerVersion represents the current version of the provider. It should be
// overwritten during the release process.
var providerVersion = "dev"

func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("HONEYCOMBIO_APIKEY", nil),
			},
			"api_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"honeycombio_query": dataSourceHoneycombioQuery(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"honeycombio_board":   newBoard(),
			"honeycombio_marker":  newMarker(),
			"honeycombio_trigger": newTrigger(),
		},
	}

	provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		config := &honeycombio.Config{
			APIKey:    d.Get("api_key").(string),
			APIUrl:    d.Get("api_url").(string),
			UserAgent: provider.UserAgent("terraform-provider-honeycombio", providerVersion),
		}
		c, err := honeycombio.NewClient(config)
		if err != nil {
			return nil, diag.FromErr(err)
		}
		return c, nil
	}

	return provider
}
