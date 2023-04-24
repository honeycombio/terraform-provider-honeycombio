# Data Source: honeycombio_column

The `honeycombio_column` data source retrieves the details of a single column in a dataset.

-> **Note** Terraform will fail unless a column is returned by the search. Ensure that your search is specific enough to return a column.
If you want to match multiple columns, use the `honeycombio_columns` data source instead.

## Example Usage

```hcl
variable "dataset" {
  type = string
}

# Retrieve the details of a single column
data "honeycombio_column" "mycol" {
  dataset = var.dataset
  name    = "mycol"
}
```

## Argument Reference

The following arguments are supported:

* `dataset` - (Required) The dataset this column is associated with
* `name` - (Required) The name of the column

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `type` - the type of the column (string, integer, float, or boolean)
* `description` - the description of the column
* `hidden` - whether or not the column is hidden from the query builder and results
* `last_written_at` - the ISO8601 formatted time that the column last received data
* `created_at` - the ISO8601 formatted time when the column was created
* `updated_at` - the  ISO8601 formatted time when the column's metadata (type, description, etc) was last changed
