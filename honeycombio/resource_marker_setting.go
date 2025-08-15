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
	"github.com/honeycombio/terraform-provider-honeycombio/honeycombio/internal/verify"
)

func newMarkerSetting() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMarkerSettingCreate,
		ReadContext:   resourceMarkerSettingRead,
		UpdateContext: resourceMarkerSettingUpdate,
		DeleteContext: resourceMarkerSettingDelete,

		Schema: map[string]*schema.Schema{
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: `The type of marker setting. (e.g. "deploy", "job-run")`,
			},
			"color": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The color set for the marker as a hex color code.",
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`), "invalid color hex code"),
			},
			"dataset": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Description:      "The dataset this marker setting belongs to. If not set, it will be Environment-wide.",
				DiffSuppressFunc: verify.SuppressEquivEnvWideDataset,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: "Timestamp when the marker setting was created.",
				Computed:    true,
				Required:    false,
				Optional:    false,
			},
			"updated_at": {
				Type:        schema.TypeString,
				Description: "Timestamp when the marker setting was last modified.",
				Computed:    true,
				Required:    false,
				Optional:    false,
			},
		},
	}
}

func resourceMarkerSettingCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	dataset := getDatasetOrAll(d)

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

func resourceMarkerSettingRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	dataset := getDatasetOrAll(d)

	var detailedErr honeycombio.DetailedError
	markerSetting, err := client.MarkerSettings.Get(ctx, dataset, d.Id())
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

func resourceMarkerSettingUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	dataset := getDatasetOrAll(d)
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

func resourceMarkerSettingDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	dataset := getDatasetOrAll(d)

	err = client.MarkerSettings.Delete(ctx, dataset, d.Id())
	if err != nil {
		return diagFromErr(err)
	}
	return nil
}
