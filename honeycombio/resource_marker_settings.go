package honeycombio

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func newMarkerSetting() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMarkerSettingCreate,
		ReadContext:   resourceMarkerSettingRead,
		UpdateContext: nil,
		DeleteContext: schema.NoopContext,

		Schema: map[string]*schema.Schema{
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"color": {
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

func resourceMarkerSettingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)

	data := &honeycombio.MarkerSetting{
		Type:  d.Get("type").(string),
		Color: d.Get("color").(string),
	}
	MarkerSetting, err := client.MarkerSettings.Create(ctx, dataset, data)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("type", MarkerSetting.Type)
	d.Set("color", MarkerSetting.Color)
	return resourceMarkerSettingRead(ctx, d, meta)
}

func resourceMarkerSettingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	markerSetting, err := client.MarkerSettings.Get(ctx, d.Get("dataset").(string), d.Id())
	if err == honeycombio.ErrNotFound {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.Set("type", markerSetting.Type)
	d.Set("color", markerSetting.Color)
	return nil
}
