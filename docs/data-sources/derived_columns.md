# Data Source: honeycombio_derived_columns

The `honeycombio_derived_columns` data source allows the derived columns of a dataset to be retrieved.

-> **API Keys** Note that this requires a [v1 API Key](https://registry.terraform.io/providers/honeycombio/honeycombio/latest/docs#v1-apis)

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

* `dataset` - (Optional) The dataset to retrieve the columns list from. If not set, an Environment-wide lookup will be performed.
* `starts_with` - (Optional) Only return derived columns starting with the given value.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `names` - a list of all the derived column names found in the dataset
