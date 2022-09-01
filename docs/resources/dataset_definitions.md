# Resource: honeycombio_dataset_definition

Updates definitions for a dataset.

-> **Note** If this dataset definitions are automatically created when a dataset is.

-> **Note** Destroying or replacing this resource will not delete the created dataset. It's not possible to delete a dataset using the API. To clear a value set it to "".

## Example Usage

```hcl
resource "honeycombio_dataset_definition" "my_dataset_definition" {
  trace_id = {
    name = "trace.trace_id"
    column_type = "column"
  }
}
```

## Argument Reference

The following arguments are supported:

-   `dataset` - (Required) Specifies which dataset to update definitions for.

### Allowed Definitions:

-   `trace_id` - Definition column for Trace ID. (default: "", column)

-> **Note** All dataset definitions can be defined by a definition column. These are standard columns or derived columns.

-   `name` - (Required) The value for the definition.
-   `column` - The column type for the of the definition.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

-   `id` - The id of the definintion.
