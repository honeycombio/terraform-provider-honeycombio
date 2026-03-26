# Honeycomb.io Provider

[Honeycomb](https://honeycomb.io) is an observability platform built for high-performance engineering teams.
Use Honeycomb to understand how your code behaves in the hands of real users and to quickly identify and resolve issues in unpredictable and highly complex cloud environments.
Honeycomb helps engineering teams spend less time chasing down mysteries and more time building fast, reliable, and great experiences for their users.

To use this provider, you must have a Honeycomb account.
[Sign up for free](http://ui.honeycomb.io/signup?&utm_source=terraform&utm_medium=partner&utm_campaign=signup&utm_keyword=&utm_content=free-product-signup) to get started.

Use the navigation to the left to read about the available resources and data sources.

## Example Usage

```hcl
terraform {
  required_providers {
    honeycombio = {
      source  = "honeycombio/honeycombio"
      version = "~> 0.47.0"
    }
  }
}

# Configure the Honeycomb provider
provider "honeycombio" {
  # You can set the API key with the environment variable HONEYCOMB_API_KEY,
  # or the HONEYCOMB_KEY_ID+HONEYCOMB_KEY_SECRET environment variable pair

  # The features block allows customization of the behavior of the Honeycomb Provider.
  # More information can be found below.
  features {}
}

variable "dataset" {
  type = string
}

# Create a marker
resource "honeycombio_marker" "hello" {
  message = "Hello world!"

  dataset = var.dataset
}
```

More advanced examples can be found in the [example directory](https://github.com/honeycombio/terraform-provider-honeycombio/tree/main/example).

## Datasets

Several resources in this provider accept a `dataset` or `datasets` argument to specify which Honeycomb Dataset the resource belongs to.
These resources include but aren't limited to:

* Queries
* Triggers
* SLOs
* Markers
* Columns
* Boards

The `dataset` and `datasets` arguments expect a Dataset **slug**, not a Dataset name or ID.
Dataset slugs appear in the URL of the Dataset in the Honeycomb UI, or in the `slug` field of the [Dataset API](https://api-docs.honeycomb.io/api/datasets/createdataset#datasets/createdataset/t=response&c=200&path=slug).

## Configuring for Honeycomb EU

If you are a Honeycomb EU customer, override the default API host using a `provider` block or the `HONEYCOMB_API_ENDPOINT` environment variable.

```hcl
provider "honeycombio" {
  api_url = "https://api.eu1.honeycomb.io"
}
```

## Authentication

The Honeycomb Terraform provider requires an API key to communicate with the Honeycomb APIs.
The provider supports both v1 and v2 APIs, each requiring a different key type.
You can configure a single provider instance with both key types, but at least one is required.

To learn more about API Keys, visit [Best Practices for API Keys](https://docs.honeycomb.io/get-started/best-practices/api-keys/).

### v1 APIs

v1 APIs require a Configuration Key.
Most resources and data sources use v1 APIs.
You can manage Configuration Key permissions in _Environment settings_.

Set the key with the `api_key` argument, or use one of these environment variables:

- `HONEYCOMB_API_KEY`
- `HONEYCOMBIO_APIKEY`

If both are set, `HONEYCOMB_API_KEY` takes priority.

~> **Note** Use only the Key value; do not include the Key ID.

### v2 APIs

v2 APIs require a Management Key.
You can manage Management Key permissions in _Team settings_.
Resources and data sources that use v2 APIs are noted individually, along with the required scope.

Set the key pair with the `api_key_id` and `api_key_secret` arguments, or use the `HONEYCOMB_KEY_ID` and `HONEYCOMB_KEY_SECRET` environment variables.
`api_key_id` takes the Key ID and `api_key_secret` takes the Key Secret.

~> **Note** Avoid hard-coding API keys in Terraform configuration.
Use the environment variables instead.

## Argument Reference

Arguments accepted by this provider include:

* `api_key` - (Optional) The Configuration Key for v1 API access. Can also be set with the `HONEYCOMB_API_KEY` or `HONEYCOMBIO_APIKEY` environment variables.
* `api_key_id` - (Optional) The Key ID portion of a Management Key for v2 API access. Can also be set with the `HONEYCOMB_KEY_ID` environment variable.
* `api_key_secret` - (Optional) The Key Secret portion of a Management Key for v2 API access. Can also be set with the `HONEYCOMB_KEY_SECRET` environment variable.
* `api_url` - (Optional) Override the Honeycomb API URL. Can also be set with `HONEYCOMB_API_ENDPOINT`. Defaults to `https://api.honeycomb.io`.
* `debug` - (Optional) Log additional debug information. To view the logs, set `TF_LOG` to at least `debug`.
* `features` - (Optional) Customize the behavior of specific Honeycomb Provider resources. See [Features Block](#features-block).

At least one of `api_key`, or the `api_key_id` and `api_key_secret` pair, must be configured.

## Features Block

The `features` block lets you modify the behavior of certain resources.
If the default behavior works for your use case, no configuration is needed.

~> **Warning** Some behaviors enabled by the features block can cause data loss.
Review each option carefully before enabling it.

### Example Usage

Each option can be configured individually.
This example shows all available options:

```hcl
provider "honeycombio" {
  features {
    column {
      import_on_conflict = true
    }
    dataset {
      import_on_conflict = true
    }
  }
}
```

### Arguments Reference

The `features` block supports:

* `column` - (Optional) A `column` block as defined below.
* `dataset` - (Optional) A `dataset` block as defined below.

---
#### `column` block

* `import_on_conflict` - (Optional) When `true`, if a column already exists, the provider imports and updates it rather then returning an error. Defaults to `false`.

~> **Warning** Changing a column type (for example, from `string` to `boolean`) can cause data loss. Use this option with caution.

---
#### `dataset` block

* `import_on_conflict` - (Optional) When `true`, if a dataset already exists, the provider imports and updates it rather than returning an error. Defaults to `false`.
