# Data Source: honeycombio_slos
The SLOs data source retrieves the SLOs of a dataset or environment, with the option of narrowing the retrieval by providing a `detail_filter`.

~> **Note** Multi-Dataset SLOs are not supported yet for this data source.


## Example Usage

```hcl
variable "dataset" {
  type = string
}

# returns all SLOs
data "honeycombio_slos" "all" {
  dataset = var.dataset
}

# only returns the SLOs starting with 'foo_'
data "honeycombio_slos" "foo" {
  dataset     = var.dataset

  detail_filter {
    name        = "name"
    value_regex = "foo_*"
  }
}
```

## Argument Reference

The following arguments are supported:

* `dataset` - (Optional) The dataset to retrieve the SLOs list from. If omitted, the lookup will be Environment-wide.
* `detail_filter` - (Optional) a block to further filter results as described below. `name` must be set when providing a filter.

To further filter the SLO results, a `detail_filter` block can be provided which accepts the following arguments:

* `name` - (Required) The name of the detail field to filter by. Currently only `name` is supported.
* `value` - (Optional) The value of the detail field to match on.
* `value_regex` - (Optional) A regular expression string to apply to the value of the detail field to match on.

~> **Note** one of `value` or `value_regex` is required.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `ids` - a list of all the SLO IDs found in the dataset
