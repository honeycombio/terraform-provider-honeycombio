package honeycombio

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func newMarker() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMarkerCreate,
		ReadContext:   resourceMarkerRead,
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

func resourceMarkerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)

	data := &honeycombio.Marker{
		Message: d.Get("message").(string),
		Type:    d.Get("type").(string),
		URL:     d.Get("url").(string),
	}
	marker, err := client.Markers.Create(ctx, dataset, data)
	if err != nil {
		return diagFromErr(err)
	}

	d.SetId(marker.ID)
	return resourceMarkerRead(ctx, d, meta)
}

func resourceMarkerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	var detailedErr honeycombio.DetailedError
	marker, err := client.Markers.Get(ctx, d.Get("dataset").(string), d.Id())
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

	d.SetId(marker.ID)
	d.Set("message", marker.Message)
	d.Set("type", marker.Type)
	d.Set("url", marker.URL)
	return nil
}
