package honeycombio

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	honeycombio "github.com/kvrhdn/go-honeycombio"
)

func newMarker() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMarkerCreate,
		ReadContext:   resourceMarkerRead,
		UpdateContext: nil,
		DeleteContext: resourceMarkerDelete,

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

func resourceMarkerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)

	data := honeycombio.MarkerCreateData{
		Message: d.Get("message").(string),
		Type:    d.Get("type").(string),
		URL:     d.Get("url").(string),
	}
	marker, err := client.Markers.Create(ctx, dataset, data)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(marker.ID)
	return resourceMarkerRead(ctx, d, meta)
}

func resourceMarkerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	marker, err := client.Markers.Get(ctx, d.Get("dataset").(string), d.Id())
	if err != nil {
		if err == honeycombio.ErrNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(marker.ID)
	d.Set("message", marker.Message)
	d.Set("type", marker.Type)
	d.Set("url", marker.URL)
	return nil
}

func resourceMarkerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// do nothing on destroy
	return nil
}
