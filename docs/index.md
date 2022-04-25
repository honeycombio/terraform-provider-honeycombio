# Honeycomb.io Provider

[Honeycomb](https://honeycomb.io) provides observability for high-performance engineering teams so they can quickly understand what their code does in the hands of real users in unpredictable and highly complex cloud environments.
Honeycomb customers stop wasting precious time on engineering mysteries because they can quickly solve them and know exactly how to create fast, reliable, and great customer experiences.

In order to use this provider, you must have a Honeycomb account. You can get started today with a [free account](http://ui.honeycomb.io/signup?&utm_source=terraform&utm_medium=partner&utm_campaign=signup&utm_keyword=&utm_content=free-product-signup).

Use the navigation to the left to read about the available resources and data sources.

## Example usage

```hcl
terraform {
  required_providers {
    honeycombio = {
      source  = "honeycombio/honeycombio"
      version = "~> 0.5.0"
    }
  }
}

# Configure the Honeycomb provider
provider "honeycombio" {
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

The Honeycomb provider requires an API key to communicate with the Honeycomb API. API keys and their permissions can be managed in _Team settings_.

The key can be set with the `api_key` argument or via the `HONEYCOMBIO_APIKEY` environment variable.

~> **Note** Hard-coding API keys in any Terraform configuration is not recommended. Consider using the `HONEYCOMBIO_APIKEY` environment variable.

## Argument Reference

Arguments accepted by this provider include:

* `api_key` - (Required) The Honeycomb API key to use. It can also be set using the `HONEYCOMBIO_APIKEY` environment variable.
* `api_url` - (Optional) Override the url of the Honeycomb.io API. Defaults to `https://api.honeycomb.io`.
* `debug` - (Optional) Enable to log additional debug information. To view the logs, set `TF_LOG` to at least debug.
