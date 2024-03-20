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
      version = "~> 0.23.0"
    }
  }
}

# Configure the Honeycomb provider
provider "honeycombio" {
  # You can set the API key with the environment variable HONEYCOMB_API_KEY
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

### Configuring the provider for Honeycomb EU

If you are a Honeycomb EU customer, to use the provider you must override the default API host.
This can be done with a `provider` block or by setting the `HONEYCOMB_API_ENDPOINT` environment variable.

```hcl
provider "honeycombio" {
  api_url = "https://api.eu1.honeycomb.io"
}
```

## Authentication

The Honeycomb provider requires an API key to communicate with the Honeycomb API. API keys and their permissions can be managed in _Team settings_.

The key can be set with the `api_key` argument or via the `HONEYCOMB_API_KEY` or `HONEYCOMBIO_APIKEY` environment variable.

`HONEYCOMB_API_KEY` environment variable will take priority over the `HONEYCOMBIO_APIKEY` environment variable.

~> **Note** Hard-coding API keys in any Terraform configuration is not recommended. Consider using the one of the environment variable options.

## Argument Reference

Arguments accepted by this provider include:

* `api_key` - (Required) The Honeycomb API key to use. It can also be set using `HONEYCOMB_API_KEY` or `HONEYCOMBIO_APIKEY` environment variables.
* `api_url` - (Optional) Override the URL of the Honeycomb.io API. It can also be set using `HONEYCOMB_API_ENDPOINT`. Defaults to `https://api.honeycomb.io`.
* `debug` - (Optional) Enable to log additional debug information. To view the logs, set `TF_LOG` to at least debug.
