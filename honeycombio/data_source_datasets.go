package honeycombio

import (
	"context"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kvrhdn/go-honeycombio"
	"github.com/kvrhdn/terraform-provider-honeycombio/honeycombio/internal/hashcode"
)

func dataSourceHoneycombioDatasets() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHoneycombioDatasetsRead,

		Schema: map[string]*schema.Schema{
			"starts_with": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"slugs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceHoneycombioDatasetsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	datasets, err := client.Datasets.List(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	value, ok := d.GetOk("starts_with")
	if ok {
		startsWith := value.(string)

		for i := len(datasets) - 1; i >= 0; i-- {
			if !strings.HasPrefix(datasets[i].Name, startsWith) {
				datasets = append(datasets[:i], datasets[i+1:]...)
			}
		}
	}

	names := make([]string, len(datasets))
	slugs := make([]string, len(datasets))

	for i, d := range datasets {
		names[i] = d.Name
		slugs[i] = d.Slug
	}

	d.Set("names", names)
	d.Set("slugs", slugs)

	d.SetId(strconv.Itoa(hashcode.String(strings.Join(names, ","))))
	return nil
}
