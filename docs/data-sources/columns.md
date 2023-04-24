# Data Source: honeycombio_columns

The columns data source allows the columns of a dataset to be retrieved.

## Example Usage

```hcl
variable "dataset" {
  type = string
}

# returns all columns
data "honeycombio_columns" "all" {
  dataset = var.dataset
}

# only returns the columns starting with 'foo_'
data "honeycombio_columns" "foo" {
  dataset     = var.dataset
  starts_with = "foo_"
}
```

## Argument Reference

The following arguments are supported:

* `dataset` - (Required) The dataset to retrieve the columns list from
* `starts_with` - (Optional) Only return columns starting with the given value.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `names` - a list of all the column names found in the dataset
