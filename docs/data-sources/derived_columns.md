# Data Source: honeycombio_derived_columns

The `honeycombio_derived_columns` data source allows the derived columns of a dataset to be retrieved.

## Example Usage

```hcl
variable "dataset" {
  type = string
}

# returns all columns
data "honeycombio_derived_columns" "all" {
  dataset = var.dataset
}

# only returns the derived columns starting with 'foo_'
data "honeycombio_derived_columns" "foo" {
  dataset     = var.dataset
  starts_with = "foo_"
}
```

## Argument Reference

The following arguments are supported:

* `dataset` - (Required) The dataset to retrieve the columns list from. Use `__all__` for Environment-wide derived columns.
* `starts_with` - (Optional) Only return derived columns starting with the given value.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `names` - a list of all the derived column names found in the dataset
