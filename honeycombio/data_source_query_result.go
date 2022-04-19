package honeycombio

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func dataSourceHoneycombioQueryResult() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHoneycombioQueryResultRead,

		Schema: map[string]*schema.Schema{
			"dataset": {
				Type:     schema.TypeString,
				Required: true,
			},
			"query_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			// outputs
			"query_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"graph_image_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"results": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
			},
		},
	}
}

func dataSourceHoneycombioQueryResultRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)
	queryID := d.Get("query_id").(string)

	queryResult, err := client.QueryResults.Create(ctx, dataset, &honeycombio.QueryResultRequest{ID: queryID})
	if err != nil {
		return diag.FromErr(err)
	}
	err = client.QueryResults.Get(ctx, dataset, queryResult)
	if err != nil {
		return diag.FromErr(err)
	}

	results := make([]map[string]string, len(queryResult.Data.Results))
	for i, qr := range queryResult.Data.Results {
		result := make(map[string]string, len(qr.Data))
		for k, v := range qr.Data {
			// convert all values to strings as the Plugin SDK can't handle dynamic types/objects.
			//
			// The not-yet-stable 'terraform-plugin-go' looks to have support for complex object results
			// so this may be improved in future
			result[k] = fmt.Sprintf("%v", v)
		}
		results[i] = result
	}

	d.SetId(queryResult.ID)
	d.Set("query_url", queryResult.Links.Url)
	d.Set("graph_image_url", queryResult.Links.GraphUrl)
	d.Set("results", results)

	return nil
}
