# Data Source: honeycombio_datasets

The Datasets data source retrieves the Environment's Datasets.

## Example Usage

```hcl
# returns all datasets
data "honeycombio_datasets" "all" {}

# only returns the datasets with names starting with 'foo_'
data "honeycombio_datasets" "foo" {
  detail_filter {
    name        = "name"
    value_regex = "foo_*"
  }
}
```

## Argument Reference

The following arguments are supported:

* `detail_filter` - (Optional) a block to further filter results as described below. `name` must be set when providing a filter. Conflicts with `starts_with`.
* `starts_with` - (Optional) Deprecated: use `detail_filter` instead. Only return datasets whose name starts with the given value.

To filter the results, a `detail_filter` block can be provided which accepts the following arguments:

* `name` - (Required) The name of the detail field to filter by. Currently only `name` is supported.
* `value` - (Optional) The value of the detail field to match on.
* `value_regex` - (Optional) A regular expression string to apply to the value of the detail field to match on.

~> **Note** one of `value` or `value_regex` is required.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `names` - a list of all the dataset names.
* `slugs` - a list of all the dataset slugs.
