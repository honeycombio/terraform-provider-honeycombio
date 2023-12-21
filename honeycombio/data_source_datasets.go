package honeycombio

import (
	"context"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/hashcode"
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
		return diagFromErr(err)
	}

	var startsWith string
	if value, ok := d.GetOk("starts_with"); ok {
		startsWith = value.(string)
	}

	names := make([]string, 0, len(datasets))
	slugs := make([]string, 0, len(datasets))
	for _, d := range datasets {
		if startsWith != "" && !strings.HasPrefix(d.Name, startsWith) {
			continue
		}
		names = append(names, d.Name)
		slugs = append(slugs, d.Slug)
	}
	d.Set("names", names)
	d.Set("slugs", slugs)

	d.SetId(strconv.Itoa(hashcode.String(strings.Join(names, ","))))
	return nil
}
