# Resource: honeycombio_dataset

Creates a Dataset in an Environment.

-> **Note**: prior to version 0.27.0 of the provider, datasets were *not* deleted on destroy but left in place and only removed from state.

## Example Usage

```hcl
resource "honeycombio_dataset" "my_dataset" {
  name        = "My dataset"
  description = "buzzing with data"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the dataset.
* `description` - (Optional) A longer description for dataset.
* `expand_json_depth` - (Optional) The maximum unpacking depth of nested JSON fields.
* `delete_protected` - (Optional) the current state of the Dataset's deletion protection status. Defaults to true. Cannot be set to false on create.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `slug` - The slug of the dataset.
* `created_at` - ISO8601-formatted time the dataset was created
* `last_written_at` - ISO8601-formatted time the dataset was last written to (received event data)

## Import

Datasets can be imported by their slug, e.g.

```shell
$ terraform import honeycombio_dataset.my_dataset my-dataset
```
