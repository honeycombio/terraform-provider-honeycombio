# Data Source: honeycombio_environment

The `honeycombio_environment` data source retrieves the details of a single Environment.
If you want to retrieve multiple Environments, use the `honeycombio_environments` data source instead.

-> **NOTE** This data source requires the provider be configured with a Management Key with `environments:read` in the configured scopes.


## Example Usage

```hcl
# Retrieve the details of an Environment
data "honeycombio_environment" "prod" {
  id = "hcaen_01j1d7t02zf7wgw7q89z3t60vf"
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Required) The ID of the Environment

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `name` - the Environment's name.
* `slug` - the Environment's slug.
* `description` - the Environment's description.
* `color` - the Environment's color.
* `delete_protected` - the current state of the Environment's deletion protection status.

