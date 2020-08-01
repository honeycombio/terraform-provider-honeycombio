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

	data := honeycombio.MarkerCreateData{
		Message: d.Get("message").(string),
		Type:    d.Get("type").(string),
		URL:     d.Get("url").(string),
	}
	marker, err := client.Markers.Create(data)
	if err != nil {
		return err
	}

	d.SetId(marker.ID)
	return resourceMarkerRead(d, meta)
}

func resourceMarkerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*honeycombio.Client)

	marker, err := client.Markers.Get(d.Id())
	if err != nil {
		if err == honeycombio.ErrNotFound {
			d.SetId("")
			return nil
		}
		return err
	}

	d.SetId(marker.ID)
	d.Set("message", marker.Message)
	d.Set("type", marker.Type)
	d.Set("url", marker.URL)
	return nil
}

func resourceMarkerDelete(d *schema.ResourceData, meta interface{}) error {
	// do nothing on destroy
	return nil
}
