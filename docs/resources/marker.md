# Resource: honeycombio_marker

Creates a marker. For more information about markers, check out [Annotate the timeline with Markers](https://docs.honeycomb.io/working-with-your-data/customizing-your-query/markers/).

-> Destroying or replacing this resource will not delete the previously created marker.
This is intentional to preserve the markers.

## Example Usage

```hcl
variable "dataset" {
  type = string
}

variable "app_version" {
  type = string
}

resource "honeycombio_marker" "marker" {
  message = "deploy ${var.app_version}"
  type    = "deploy"
  url     = "http://www.example.com/"

  dataset = var.dataset
}
```

## Argument Reference

The following arguments are supported:

* `dataset` - (Required) The dataset this marker is placed on. Use `__all__` for Environment-wide markers.
* `message` - (Optional) The message on the marker.
* `type` - (Optional) The type of the marker, Honeycomb.io can display markers in different colors depending on their type.
* `url` - (Optional) A target for the Marker. If you click on the Marker text, it will take you to this URL.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the marker.
