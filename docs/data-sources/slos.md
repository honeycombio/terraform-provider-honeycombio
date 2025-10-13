# Data Source: honeycombio_slos
The SLOs data source retrieves the SLOs of a dataset or environment, with the option of narrowing the retrieval by providing a `detail_filter`.

-> **API Keys** Note that this requires a [v1 API Key](https://registry.terraform.io/providers/honeycombio/honeycombio/latest/docs#v1-apis)

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

  detail_filter {
    name     = "tags"
    operator = "contains"
    value    = "team:core"
  }
}
```

## Argument Reference

The following arguments are supported:

* `dataset` - (Optional) The dataset to retrieve the SLOs list from. If omitted, the lookup will be Environment-wide.
* `detail_filter` - (Optional) a block to further filter results as described below. Multiple `detail_filter` blocks can be provided to filter by multiple fields. Multiple filters are combined with a logical `AND` operation, meaning all conditions must be satisfied for an SLO to be included in the results.

To further filter the SLO results, a `detail_filter` block can be provided which accepts the following arguments:

* `name` - (Required) The name of the detail field to filter by. This field must match a schema attribute of the `honeycombio_slo` resource (e.g., `name`, `tags`, `description`).
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

~> **Note:** Either `value` or `value_regex` must be specified, but not both.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `ids` - a list of all the SLO IDs found in the dataset
