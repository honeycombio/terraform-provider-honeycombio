# Honeycomb.io Provider

The Honeycomb.io provider is used to manage various resources on Honeycomb. You can use it to create and manage markers and triggers.

Use the navigation to the left to read about the available resources and data sources.

## Example usage

```hcl
# Configure the Honeycomb.io provider
provider "honeycombio" {
  version = "~> 0.1.0"

  # You can set the API key with the environment variable HONEYCOMBIO_APIKEY
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

## Authentication

The Honeycomb.io provider requires an API key to communicate with the Honeycomb.io API. API keys and their permissions can be managed in _Team settings_.

The key can be set with the `api_key` argument or via the `HONEYCOMBIO_APIKEY` environment variable.

~> **Note** Hard-coding API keys in any Terraform configuration is not recommended. Consider using the `HONEYCOMBIO_APIKEY` environment variable.

## Argument Reference

Arguments accepted by this provider include:

* `api_key` - (Required) The Honeycomb API key to use. It can also be set using the `HONEYCOMBIO_APIKEY` environment variable.
* `api_url` - (Optional) Override the url of the Honeycomb.io API. Defaults to `https://api.honeycomb.io`.
* `debug` - (Optional) Enable to log additional debug information. To view the logs, set `TF_LOG` to at least debug.
