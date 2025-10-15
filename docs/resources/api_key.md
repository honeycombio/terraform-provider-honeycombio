# Resource: honeycombio_api_key

Creates a Honeycomb API Key.
For more information about API Keys, check out [Best Practices for API Keys](https://docs.honeycomb.io/get-started/best-practices/api-keys/).

-> **API Keys** Note that this requires a [v2 API Key](https://registry.terraform.io/providers/honeycombio/honeycombio/latest/docs#v2-apis)

-> This resource requires the provider be configured with a Management Key with `api-keys:write` in the configured scopes.

## Example Usage

```hcl
resource "honeycombio_api_key" "prod_ingest" {
  name = "Production Ingest"
  type = "ingest"

  environment_id = var.environment_id

  permissions {
    create_datasets = true
  }
}

output "ingest_key" {
  value = "${honeycomb_api_key.prod_ingest.key}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the API key.
* `type` - (Required) The type of API key. Currently only `ingest` is supported.
* `environment_id` - (Required) The Environment ID the API key is scoped to.
* `disabled` - (Optional) Whether the API key is disabled. Defaults to `false`.
* `permissions` - (Optional) A configuration block (described below) setting what actions the API key can perform.

Each API key configuration may contain a single `permissions` block, which accepts the following arguments:

* `create_datasets` - (Optional) Allow this key to create missing datasets when sending telemetry. Defaults to `false`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the API Key.
* `secret` - The secret portion of the API Key.
* `key` - The API key formatted for use based on its type.

## Import

API Keys cannot be imported.
