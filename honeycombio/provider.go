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
			"dataset": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("HONEYCOMBIO_DATASET", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"honeycombio_marker":  newMarker(),
			"honeycombio_trigger": newTrigger(),
		},
		ConfigureFunc: Configure,
	}
}

func Configure(d *schema.ResourceData) (interface{}, error) {
	apiKey := d.Get("api_key").(string)
	dataset := d.Get("dataset").(string)

	return honeycombio.NewClient(apiKey, dataset), nil
}
