# Resource: honeycombio_dataset

Creates a dataset.

-> **Note** If this dataset already exists, creating this resource is a no-op.

-> **Note** Destroying or replacing this resource will not delete the created dataset. It's not possible to delete a dataset using the API.

## Example Usage

```hcl
resource "honeycombio_dataset" "my_dataset" {
  name = "My dataset"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the dataset.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `slug` - The slug of the dataset.
