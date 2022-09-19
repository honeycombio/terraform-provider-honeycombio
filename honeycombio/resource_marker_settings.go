package honeycombio

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func newMarkerType() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMarkerTypeCreate,
		ReadContext:   resourceMarkerTypeRead,
		UpdateContext: nil,
		DeleteContext: schema.NoopContext,

		Schema: map[string]*schema.Schema{
			"message": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"url": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"dataset": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceMarkerTypeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)

	data := &honeycombio.MarkerType{
		Type:  d.Get("type").(string),
		Color: d.Get("color").(string),
	}
	MarkerType, err := client.MarkerTypes.Create(ctx, dataset, data)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(MarkerType.ID)
	return resourceMarkerTypeRead(ctx, d, meta)
}

func resourceMarkerTypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	MarkerType, err := client.MarkerTypes.List(ctx, d.Get("dataset").(string))
	if err == honeycombio.ErrNotFound {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.Set("type", append(MarkerType))
	return nil
}
