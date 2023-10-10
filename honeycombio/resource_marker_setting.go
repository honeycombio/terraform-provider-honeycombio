package honeycombio

import (
	"context"
	"errors"
	"regexp"
	"time"

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
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`), "invalid color hex code"),
			},
			"dataset": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
				Required: false,
				Optional: false,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
				Required: false,
				Optional: false,
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
		return diagFromErr(err)
	}

	d.SetId(markerSetting.ID)
	d.Set("type", markerSetting.Type)
	d.Set("color", markerSetting.Color)
	return resourceMarkerSettingRead(ctx, d, meta)
}

func resourceMarkerSettingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	var detailedErr honeycombio.DetailedError
	markerSetting, err := client.MarkerSettings.Get(ctx, d.Get("dataset").(string), d.Id())
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

	d.SetId(markerSetting.ID)
	d.Set("type", markerSetting.Type)
	d.Set("color", markerSetting.Color)
	d.Set("created_at", markerSetting.CreatedAt.UTC().Format(time.RFC3339))
	d.Set("updated_at", markerSetting.UpdatedAt.UTC().Format(time.RFC3339))
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
		return diagFromErr(err)
	}

	d.SetId(markerSetting.ID)
	return resourceMarkerSettingRead(ctx, d, meta)
}

func resourceMarkerSettingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*honeycombio.Client)

	dataset := d.Get("dataset").(string)

	err := client.MarkerSettings.Delete(ctx, dataset, d.Id())
	if err != nil {
		return diagFromErr(err)
	}
	return nil
}
