# Resource: honeycombio_marker_setting

Creates a marker setting. For more information about marker settings, check out the [Marker Settings API](https://docs.honeycomb.io/api/marker-settings/).

## Example Usage

```hcl
variable "dataset" {
  type = string
}

variable "type" {
  type = string
}

resource "honeycombio_marker_setting" "markerSetting" {
  type =  var.type
  color = "#DF4661"
  dataset = var.dataset
}
```

## Argument Reference

The following arguments are supported:

* `dataset` - (Required) The dataset this marker setting is placed on.
* `type` - (Required) The type of the marker setting, Honeycomb.io can display markers in different colors depending on their type.
* `color` - (Optional) The color set for the marker.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the marker setting.
* `created_at` - Timestamp when the marker setting was created.
* `updated_at` - Timestamp when the marker setting was last modified.
