package honeycombio

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/honeycombio/internal/verify"
)

func dataSourceHoneycombioQueryResult() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHoneycombioQueryResultRead,

		Schema: map[string]*schema.Schema{
			"dataset": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The dataset to query. If not specified, an Environment-wide query will be run.",
			},
			"query_json": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateQueryJSON(),
				DiffSuppressFunc: verify.SuppressEquivJSONDiffs,
			},
			// outputs
			"query_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
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
	var querySpec honeycombio.QuerySpec
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}
	dataset := getDatasetOrAll(d)

	queryJSON, ok := d.Get("query_json").(string)
	if !ok {
		return diag.Errorf("query_json must be a string")
	}
	err = json.Unmarshal([]byte(queryJSON), &querySpec)
	if err != nil {
		return diag.FromErr(err)
	}
	query, err := client.Queries.Create(ctx, dataset, &querySpec)
	if err != nil {
		return diagFromErr(err)
	}

	queryResult, err := client.QueryResults.Create(ctx, dataset, &honeycombio.QueryResultRequest{ID: *query.ID})
	if err != nil {
		return diagFromErr(err)
	}
	err = client.QueryResults.Get(ctx, dataset, queryResult)
	if err != nil {
		return diagFromErr(err)
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
	encodedQueryJSON, err := query.Encode()
	if err != nil {
		return diag.FromErr(err)
	}
	_ = d.Set("query_json", encodedQueryJSON)
	_ = d.Set("query_id", *query.ID)
	_ = d.Set("query_url", queryResult.Links.Url)
	_ = d.Set("graph_image_url", queryResult.Links.GraphUrl)
	_ = d.Set("results", results)

	return nil
}
