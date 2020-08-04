package honeycombio

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	honeycombio "github.com/kvrhdn/go-honeycombio"
)

func Provider() *schema.Provider {
	return &schema.Provider{
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
            "honeycombio_board":  newBoard(),
			"honeycombio_marker":  newMarker(),
			"honeycombio_trigger": newTrigger(),
		},
		ConfigureFunc: Configure,
	}
}

func Configure(d *schema.ResourceData) (interface{}, error) {
	config := &honeycombio.Config{
		APIKey:    d.Get("api_key").(string),
		APIUrl:    d.Get("api_url").(string),
		UserAgent: "terraform-provider-honeycombio",
	}
	return honeycombio.NewClient(config)
}
