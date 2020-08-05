# Resource: honeycombio_marker

Creates a marker. For more information about markers, check out [Annotate the timeline with Markers](https://docs.honeycomb.io/working-with-your-data/customizing-your-query/markers/).

-> **Note** Destroying or replacing this resource will not delete the previously created marker. This is intentional to preserve the markers. At this time, it is not possible to remove markers using this provider.

## Example Usage

```hcl
variable "app_version" {
    type = string
}

resource "honeycombio_marker" "marker" {
  message = "deploy ${var.app_version}"
  type    = "deploys"
  url     = "http://www.example.com/"

  dataset = "<your dataset>
}
```

## Argument Reference

The following arguments are supported:

* `message` - (Optional) The message on the marker.
* `type` - (Optional) The type of the marker, Honeycomb.io can display markers in different colors depending on their type.
* `url` - (Optional) A link to click on.
* `dataset` - (Required) The dataset this marker is placed on.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID for the marker.
