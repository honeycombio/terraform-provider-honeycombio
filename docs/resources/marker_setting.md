# Resource: honeycombio_marker_setting

Creates a marker setting.
For more information on marker settings, check out the [Marker Settings API](https://docs.honeycomb.io/api/marker-settings/).

## Example Usage

```hcl
variable "dataset" {
  type = string
}
resource "honeycombio_marker_setting" "deploy_marker" {
  type    =  "deploy"
  color   = "#DF4661"
  dataset = var.dataset
}
```

## Argument Reference

The following arguments are supported:

* `dataset` - (Optional) The dataset this marker setting belongs to. If not set, the marker setting will be Environment-wide.
* `type` - (Required) The type of the marker setting. (e.g. "deploy", "job-run")
* `color` - (Required) The color set for the marker as a hex color code.(e.g. `#DF4661`)

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the marker setting.
* `created_at` - Timestamp when the marker setting was created.
* `updated_at` - Timestamp when the marker setting was last modified.
