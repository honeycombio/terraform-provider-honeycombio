# Resource: honeycombio_dataset

Creates a dataset.

-> **Note** If this dataset already exists, creating this resource is a no-op.

-> **Note** Destroying or replacing this resource will not delete the created dataset. It's not possible to delete a dataset using the API.

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

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `slug` - The slug of the dataset.
* `created_at` - ISO8601 formatted time the column was created
* `last_written_at` - ISO8601 formatted time the column was last written to (received event data)

## Import

Datasets can be imported by their slug, e.g.

```shell
$ terraform import honeycombio_column.my_dataset my-dataset
```

You can find the slug in the URL bar when visiting the Dataset from the UI.
