package honeycombio

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kvrhdn/go-honeycombio"
)

func newDataset() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatasetCreate,
		ReadContext:   resourceDatasetRead,
		UpdateContext: nil,
		DeleteContext: schema.NoopContext,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"slug": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDatasetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	data := &honeycombio.Dataset{
		Name: d.Get("name").(string),
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
	if err != nil {
		if err == honeycombio.ErrNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(dataset.Slug)
	d.Set("name", dataset.Name)
	d.Set("slug", dataset.Slug)
	return nil
}
