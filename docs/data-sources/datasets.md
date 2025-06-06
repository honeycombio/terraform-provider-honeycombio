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

* `detail_filter` - (Optional) a block to further filter results as described below. Multiple `detail_filter` blocks can be provided to filter by multiple fields. Multiple filters are combined with a logical `AND` operation, meaning all conditions must be satisfied for a dataset to be included in the results. Conflicts with `starts_with`.
* `starts_with` - (Optional) Deprecated: use `detail_filter` instead. Only return datasets whose name starts with the given value.

To filter the results, a `detail_filter` block can be provided which accepts the following arguments:

* `name` - (Required) The name of the detail field to filter by. This field must match a schema attribute of the `honeycombio_dataset` resource (e.g., `name`, `tags`, `description`).
* `operator` - (Optional) The comparison operator to use for filtering. Defaults to `equals`. Valid operators include:
  * `equals`, `=`, `eq` - Exact match comparison
  * `not-equals`, `!=`, `ne` - Inverse exact match comparison
  * `contains`, `in` - Substring inclusion check
  * `does-not-contain`, `not-in` - Inverse substring inclusion check
  * `starts-with` - Prefix matching
  * `does-not-start-with` - Inverse prefix matching
  * `ends-with` - Suffix matching
  * `does-not-end-with` - Inverse suffix matching
  * `>`, `gt` - Numeric greater than comparison
  * `>=`, `ge` - Numeric greater than or equal comparison
  * `<`, `lt` - Numeric less than comparison
  * `<=`, `le` - Numeric less than or equal comparison
  * `does-not-exist` - Field absence check
* `value` - (Optional) The value of the detail field to match on.
* `value_regex` - (Optional) A regular expression string to apply to the value of the detail field to match on.

~> **Note** one of `value` or `value_regex` is required.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `names` - a list of all the dataset names.
* `slugs` - a list of all the dataset slugs.
