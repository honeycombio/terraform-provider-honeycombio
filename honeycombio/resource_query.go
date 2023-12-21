package honeycombio

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/honeycombio/internal/verify"
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
				DiffSuppressFunc: verify.SupressEquivQuerySpecDiff,
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
		return diagFromErr(err)
	}

	d.SetId(*query.ID)
	return resourceQueryRead(ctx, d, meta)
}

func resourceQueryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)

	var detailedErr honeycombio.DetailedError
	query, err := client.Queries.Get(ctx, dataset, d.Id())
	if errors.As(err, &detailedErr) {
		if detailedErr.IsNotFound() {
			d.SetId("")
			return nil
		} else {
			return diagFromDetailedErr(detailedErr)
		}
	} else if err != nil {
		return diag.FromErr(err)
	}

	// Track ID at the Resource level
	d.SetId(*query.ID)
	query.ID = nil

	queryJSON, err := encodeQuery(query)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("query_json", queryJSON)
	return nil
}
