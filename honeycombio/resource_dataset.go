package honeycombio

import (
	"context"
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
		DeleteContext: schema.NoopContext,

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
	client := meta.(*honeycombio.Client)

	name := d.Get("name").(string)
	description := d.Get("description").(string)
	expandJSONDepth := d.Get("expand_json_depth").(int)
	data := &honeycombio.Dataset{
		Name:            name,
		Description:     description,
		ExpandJSONDepth: expandJSONDepth,
	}
	dataset, err := client.Datasets.Create(ctx, data)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(dataset.Slug)
	return resourceDatasetRead(ctx, d, meta)
}

func resourceDatasetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset, err := client.Datasets.Get(ctx, d.Id())
	if err == honeycombio.ErrNotFound {
		d.SetId("")
		return nil
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
