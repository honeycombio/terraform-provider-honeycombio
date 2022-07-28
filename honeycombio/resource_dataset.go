package honeycombio

import (
	"context"

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

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"description": {
				Type:         schema.TypeString,
				Computed:     false,
				Required:     false,
				Optional:     true,
				ForceNew:     false,
				ValidateFunc: validation.StringLenBetween(1, 1023),
			},
			"slug": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expand_json_depth": {
				Type:         schema.TypeInt,
				Computed:     false,
				Required:     false,
				Optional:     true,
				ForceNew:     false,
				ValidateFunc: validation.IntBetween(0, 10),
			},
		},
	}
}

func resourceDatasetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	description := d.Get("description").(string)
	expandJSONDepth := d.Get("expand_json_depth").(int)
	data := &honeycombio.Dataset{
		Name:            d.Get("name").(string),
		Description:     &description,
		ExpandJSONDepth: &expandJSONDepth,
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
	return nil
}

func resourceDatasetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	description := d.Get("description").(string)
	expandJSONDepth := d.Get("expand_json_depth").(int)
	data := &honeycombio.Dataset{
		Name:            d.Get("name").(string),
		Description:     &description,
		ExpandJSONDepth: &expandJSONDepth,
	}
	dataset, err := client.Datasets.Create(ctx, data)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(dataset.Slug)
	return resourceDatasetRead(ctx, d, meta)
}
