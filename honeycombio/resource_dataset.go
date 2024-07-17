package honeycombio

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func newDataset() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatasetCreate,
		ReadContext:   resourceDatasetRead,
		UpdateContext: resourceDatasetUpdate,
		DeleteContext: schema.NoopContext,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"slug": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expand_json_depth": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 10),
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
				Required: false,
				Optional: false,
			},
			"last_written_at": {
				Type:     schema.TypeString,
				Computed: true,
				Required: false,
				Optional: false,
			},
		},
	}
}

func resourceDatasetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	data := &honeycombio.Dataset{
		Name:            d.Get("name").(string),
		Description:     d.Get("description").(string),
		ExpandJSONDepth: d.Get("expand_json_depth").(int),
	}
	dataset, err := client.Datasets.Create(ctx, data)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(dataset.Slug)
	return resourceDatasetRead(ctx, d, meta)
}

func resourceDatasetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	var detailedErr honeycombio.DetailedError
	dataset, err := client.Datasets.Get(ctx, d.Id())
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

	d.SetId(dataset.Slug)
	d.Set("name", dataset.Name)
	d.Set("description", dataset.Description)
	d.Set("expand_json_depth", dataset.ExpandJSONDepth)
	d.Set("slug", dataset.Slug)
	d.Set("created_at", dataset.CreatedAt.UTC().Format(time.RFC3339))
	d.Set("last_written_at", dataset.CreatedAt.UTC().Format(time.RFC3339))
	return nil
}

func resourceDatasetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	data := &honeycombio.Dataset{
		Name:            d.Get("name").(string),
		Description:     d.Get("description").(string),
		ExpandJSONDepth: d.Get("expand_json_depth").(int),
	}
	dataset, err := client.Datasets.Update(ctx, data)
	if err != nil {
		return diagFromErr(err)
	}

	d.SetId(dataset.Slug)
	return resourceDatasetRead(ctx, d, meta)
}
