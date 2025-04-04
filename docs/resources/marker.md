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

resource "honeycombio_marker" "app_deploy" {
  message = "deploy ${var.app_version}"
  type    = "deploy"
  url     = "http://www.example.com/"

  dataset = var.dataset
}
```

## Argument Reference

The following arguments are supported:

* `dataset` - (Optional) The dataset where this marker is placed. If not set, an Environment-wide Marker will be created.
* `message` - (Optional) A message that appears above the marker and can be used to describe the marker.
* `type` - (Optional) The type of the marker (e.g. "deploy", "job-run")
* `url` - (Optional) A target URL for the Marker. Rendered as a link in the UI..

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the marker.
