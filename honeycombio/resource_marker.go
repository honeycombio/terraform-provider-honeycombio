package honeycombio

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	honeycombio "github.com/kvrhdn/go-honeycombio"
)

func newMarker() *schema.Resource {
	return &schema.Resource{
		Create: resourceMarkerCreate,
		Read:   resourceMarkerRead,
		Update: nil,
		Delete: resourceMarkerDelete,

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
		},
	}
}

func resourceMarkerCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*honeycombio.Client)

	data := honeycombio.CreateMarkerData{
		Message: d.Get("message").(string),
		Type:    d.Get("type").(string),
		URL:     d.Get("url").(string),
	}
	marker, err := client.CreateMarker(data)
	if err != nil {
		return err
	}

	d.SetId(marker.ID)
	return resourceMarkerRead(d, meta)
}

func resourceMarkerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*honeycombio.Client)

	markers, err := client.ListMarkers()
	if err != nil {
		return err
	}

	marker := findMarkerWithID(markers, d.Id())
	if marker == nil {
		d.SetId("")
		return nil
	}

	d.SetId(marker.ID)
	d.Set("message", marker.Message)
	d.Set("type", marker.Type)
	d.Set("url", marker.URL)
	return nil
}

func findMarkerWithID(markers []honeycombio.Marker, id string) *honeycombio.Marker {
	for _, m := range markers {
		if m.ID == id {
			return &m
		}
	}
	return nil
}

func resourceMarkerDelete(d *schema.ResourceData, c interface{}) error {
	// do nothing on destroy
	return nil
}
