# Data Source: honeycombio_environment

The `honeycombio_environment` data source retrieves the details of a single Environment.

-> **API Keys** Note that this requires a [v2 API Key](https://registry.terraform.io/providers/honeycombio/honeycombio/latest/docs#v2-apis)

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
  detail_filter {
    name  = "name"
    value = "Classic"
  }
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
* `detail_filter` - (Optional) a block to further filter results as described below. Multiple `detail_filter` blocks can be provided to filter by multiple fields. Multiple filters are combined with a logical `AND` operation, meaning all conditions must be satisfied for an environment to be included in the results.

To filter the results, a `detail_filter` block can be provided which accepts the following arguments:

* `name` - (Required) The name of the detail field to filter by. This field must match a schema attribute of the `honeycombio_environment` resource (e.g., `name`, `color`, `description`).
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

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `name` - the Environment's name.
* `slug` - the Environment's slug.
* `description` - the Environment's description.
* `color` - the Environment's color.
* `delete_protected` - the current state of the Environment's deletion protection status.
