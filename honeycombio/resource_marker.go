package honeycombio

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/honeycombio/internal/verify"
)

func newMarker() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMarkerCreate,
		ReadContext:   resourceMarkerRead,
		UpdateContext: nil,
		DeleteContext: schema.NoopContext,

		CustomizeDiff: resourceMarkerCustomizeDiff,

		Schema: map[string]*schema.Schema{
			"message": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: `A message that appears above the marker and can be used to describe the marker.`,
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: `The type of the marker (e.g. "deploy", "job-run")`,
			},
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "A target URL for the Marker. Rendered as a link in the UI.",
			},
			"start_time": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(1),
				Description:  "The time the marker is placed at, in Unix Time (seconds since epoch). Defaults to the marker's creation time if not set. Changing this creates a new marker; the previous one is retained.",
			},
			"end_time": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(1),
				Description:  "The end time of the marker, in Unix Time (seconds since epoch). Used to create a time-range marker: requires `start_time` to be set and must be greater than or equal to it. Changing this creates a new marker; the previous one is retained.",
			},
			"dataset": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Description:      "The dataset where this marker is placed. If not set, it will be Environment-wide.",
				DiffSuppressFunc: verify.SuppressEquivEnvWideDataset,
			},
		},
	}
}

func resourceMarkerCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
	if cfg := d.GetRawConfig(); !cfg.IsNull() {
		startSet := !cfg.GetAttr("start_time").IsNull()
		endSet := !cfg.GetAttr("end_time").IsNull()
		if endSet && !startSet {
			return errors.New("end_time cannot be set without start_time")
		}
	}

	start := d.Get("start_time").(int)
	end := d.Get("end_time").(int)
	if start > 0 && end > 0 && end < start {
		return fmt.Errorf("end_time (%d) must be greater than or equal to start_time (%d)", end, start)
	}
	return nil
}

func resourceMarkerCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	dataset := getDatasetOrAll(d)

	data := &honeycombio.Marker{
		Message:   d.Get("message").(string),
		Type:      d.Get("type").(string),
		URL:       d.Get("url").(string),
		StartTime: int64(d.Get("start_time").(int)),
		EndTime:   int64(d.Get("end_time").(int)),
	}
	marker, err := client.Markers.Create(ctx, dataset, data)
	if err != nil {
		return diagFromErr(err)
	}

	d.SetId(marker.ID)
	return resourceMarkerRead(ctx, d, meta)
}

func resourceMarkerRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	dataset := getDatasetOrAll(d)

	var detailedErr honeycombio.DetailedError
	marker, err := client.Markers.Get(ctx, dataset, d.Id())
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
	d.Set("start_time", marker.StartTime)
	d.Set("end_time", marker.EndTime)
	return nil
}
