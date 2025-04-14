# Data Source: honeycombio_environment

The `honeycombio_environment` data source retrieves the details of a single Environment.

~> **Warning** Terraform will fail unless exactly one environment is returned by the search.
  Ensure that your search is specific enough to return a single environment only.
  If you want to retrieve multiple environments, use the `honeycombio_environments` data source instead.

-> This data source requires the provider be configured with a Management Key with `environments:read` in the configured scopes.


## Example Usage

```hcl
# Retrieve the details of an Environment
data "honeycombio_environment" "prod" {
  id = "hcaen_01j1d7t02zf7wgw7q89z3t60vf"
}
```

### Filter Example

```hcl
data "honeycombio_environment" "classic" {
  detail_filter = "name"
  value         = "Classic"
}

data "honeycombio_environment" "prod" {
  detail_filter {
    name  = "name"
    value = "prod"
  }
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Optional) The ID of the Environment. Conflicts with `detail_filter`.
* `detail_filter` - (Optional) a block to further filter results as described below. `name` must be set when providing a filter.

To filter the results, a `detail_filter` block can be provided which accepts the following arguments:

* `name` - (Required) The name of the detail field to filter by. Currently only `name` is supported.
* `value` - (Optional) The value of the detail field to match on.
* `value_regex` - (Optional) A regular expression string to apply to the value of the detail field to match on.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `name` - the Environment's name.
* `slug` - the Environment's slug.
* `description` - the Environment's description.
* `color` - the Environment's color.
* `delete_protected` - the current state of the Environment's deletion protection status.
