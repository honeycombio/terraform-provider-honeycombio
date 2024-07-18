# Data Source: honeycombio_environments

The Environments data source retrieves the Team's environments.

-> **NOTE** This data source requires the provider be configured with a Management Key with `environments:read` in the configured scopes.

## Example Usage

```hcl
# returns all Environments
data "honeycombio_environments" "all" {}

# only returns the Environments starting with 'foo_'
data "honeycombio_environments" "foo" {
  detail_filter {
    name        = "name"
    value_regex = "foo_*"
  }
}
```

## Argument Reference

The following arguments are supported:

* `detail_filter` - (Optional) a block to further filter results as described below. `name` must be set when providing a filter.

To filter the results, a `detail_filter` block can be provided which accepts the following arguments:

* `name` - (Required) The name of the detail field to filter by. Currently only `name` is supported.
* `value` - (Optional) The value of the detail field to match on.
* `value_regex` - (Optional) A regular expression string to apply to the value of the detail field to match on.

~> **Note** one of `value` or `value_regex` is required.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `ids` - a list of all the Environment IDs found in the Team.
