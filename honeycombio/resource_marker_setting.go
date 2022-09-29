package honeycombio

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func newMarkerSetting() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMarkerSettingCreate,
		ReadContext:   resourceMarkerSettingRead,
		UpdateContext: resourceMarkerSettingUpdate,
		DeleteContext: resourceMarkerSettingDelete,

		Schema: map[string]*schema.Schema{
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"color": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`), "invalide color hex code"),
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
	markerSetting, err := client.MarkerSettings.Create(ctx, dataset, data)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(markerSetting.ID)
	d.Set("type", markerSetting.Type)
	d.Set("color", markerSetting.Color)
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

	d.SetId(markerSetting.ID)
	d.Set("type", markerSetting.Type)
	d.Set("color", markerSetting.Color)
	return nil
}

func resourceMarkerSettingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)
	markerType := d.Get("type").(string)
	color := d.Get("color").(string)

	data := &honeycombio.MarkerSetting{
		Type:  markerType,
		Color: color,
	}
	markerSetting, err := client.MarkerSettings.Update(ctx, dataset, data)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(markerSetting.ID)
	return resourceMarkerSettingRead(ctx, d, meta)
}

func resourceMarkerSettingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)

	err := client.MarkerSettings.Delete(ctx, dataset, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
