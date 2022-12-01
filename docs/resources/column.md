# Resource: honeycombio_column

Creates a column in a dataset.

-> **Note** Destroying or replacing this resource will not delete the previously created column. This is intentional to preserve the column and its data. The API supports deleting columns, so if this is required for your use case please open an issue in [honeycombio/terraform-provider-honeycombio](https://github.com/honeycombio/terraform-provider-honeycombio).

## Example Usage

```hcl
variable "dataset" {
  type = string
}

resource "honeycombio_column" "duration_ms" { 
  name        = "duration_ms_log10"
  type        = "float"
  description = "Duration of the trace"

  dataset = var.dataset
}
```

## Argument Reference

The following arguments are supported:

* `dataset` - (Required) The dataset this column is added to.
* `name` - (Required) The name of the column. Must be unique per dataset.
* `key_name` - (Deprecated) Please use `name` instead. The name of the column. Must be unique per dataset. Conficts with `name`.
* `type` - (Optional) The type of the column, allowed values are `string`, `float`, `integer` and `boolean`. Defaults to `string`.
* `hidden` - (Optional) Whether this column should be hidden in the query builder and sample data. Defaults to false.
* `description` - (Optional) A description that is shown in the UI.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the column.
* `created_at` - ISO8601 formatted time the column was created
* `updated_at` - ISO8601 formatted time the column was updated
* `last_written_at` - ISO8601 formatted time the column was last written to (received event data)

## Import

Columns can be imported using a combination of the dataset name and their name, e.g.

```
$ terraform import honeycombio_column.my_column my-dataset/duration_ms
```
