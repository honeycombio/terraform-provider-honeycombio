package honeycombio

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/honeycombio/internal/verify"
)

func newMarker() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMarkerCreate,
		ReadContext:   resourceMarkerRead,
		UpdateContext: nil,
		DeleteContext: schema.NoopContext,

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

func resourceMarkerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getConfiguredClient(meta)
	if err != nil {
		return diagFromErr(err)
	}

	dataset := getDatasetOrAll(d)

	message, ok := d.Get("message").(string)
	if !ok {
		return diag.Errorf("message must be a string")
	}
	typeStr, ok := d.Get("type").(string)
	if !ok {
		return diag.Errorf("type must be a string")
	}
	url, ok := d.Get("url").(string)
	if !ok {
		return diag.Errorf("url must be a string")
	}
	data := &honeycombio.Marker{
		Message: message,
		Type:    typeStr,
		URL:     url,
	}
	marker, err := client.Markers.Create(ctx, dataset, data)
	if err != nil {
		return diagFromErr(err)
	}

	d.SetId(marker.ID)
	return resourceMarkerRead(ctx, d, meta)
}

func resourceMarkerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	_ = d.Set("message", marker.Message)
	_ = d.Set("type", marker.Type)
	_ = d.Set("url", marker.URL)
	return nil
}
