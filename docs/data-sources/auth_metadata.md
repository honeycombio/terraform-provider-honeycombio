# Data Source: honeycombio_auth_metadata

The `honeycombio_auth_metadata` data source retreives information about the API key used to authenticate the provider.

## Example Usage

```hcl
data "honeycombio_auth_metadata" "current" {}

output "team_name" {
  value = honeycombio_auth_metadata.current.team.name
}

output "environment_slug" {
  value = honeycombio_auth_metadata.current.environment.slug
} 

output "slo_management_access" {
  value = honeycombio_auth_metadata.current.api_key_access.slos
}
```

## Argument Reference

There are no arguments available for this data source.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `api_key_access` - The authorizations granted for the API key used to authenticate the provider.
  See [the documentation](https://docs.honeycomb.io/working-with-your-data/settings/api-keys/) for more information.
  * `boards` - `true` if this API key can create and manage Boards.
  * `columns` - `true` if this API key can create and manage can create and manage Queries, Columns, Derived Columns, and Query Annotations
  * `datasets` - `true` if this API key can create and manage Datasets.
  * `events` - `true` if this API key can key can send events to Honeycomb.
  * `markers` - `true` if this API key can create and manage Markers.
  * `queries` - `true` if this API key can execute existing Queries via the Query Data API.
  * `recipients` - `true` if this API key can create and manage Recipients.
  * `slos` - `true` if this API key can create and manage SLOs.
  * `triggers` - `true` if this API key can create and manage Triggers.
* `environment` - Information about the Environment the API key is scoped to.
  * `classic` - `true` if this API key belongs to a [Honeycomb Classic](https://docs.honeycomb.io/honeycomb-classic/) environment.
  * `name` - The name of the Environment. For Classic environments, this will be null.
  * `slug` - The slug of the Environment. For Classic environments, this will be null.
* `team` - Information about the Team the API key is scoped to.
  * `name` - The name of the Team.
  * `slug` - The slug of the Team.
