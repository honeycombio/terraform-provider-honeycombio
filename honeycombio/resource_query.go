package honeycombio

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kvrhdn/go-honeycombio"
)

func newQuery() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceQueryCreate,
		ReadContext:   resourceQueryRead,
		UpdateContext: nil,
		DeleteContext: schema.NoopContext,
		Importer:      nil,

		Schema: map[string]*schema.Schema{
			"query_json": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateQueryJSON(),
			},
			"dataset": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceQueryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)
	var querySpec honeycombio.QuerySpec
	err := json.Unmarshal([]byte(d.Get("query_json").(string)), &querySpec)
	if err != nil {
		return diag.FromErr(err)
	}

	query, err := client.Queries.Create(ctx, dataset, &querySpec)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*query.ID)
	return resourceQueryRead(ctx, d, meta)
}

func resourceQueryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)

	query, err := client.Queries.Get(ctx, dataset, d.Id())
	if err != nil {
		if err == honeycombio.ErrNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(*query.ID)
	// Normalize the query before converting it to JSON to avoid unwanted diffs
	query.ID = nil
	if query.FilterCombination == "" {
		query.FilterCombination = honeycombio.FilterCombinationAnd
	}
	queryJSON, err := encodeQuery(query)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("query_json", queryJSON)
	return nil
}
