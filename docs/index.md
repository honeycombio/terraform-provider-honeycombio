# Honeycomb.io Provider

[Honeycomb](https://honeycomb.io) provides observability for high-performance
engineering teams so they can quickly understand what their code does in the hands
of real users in unpredictable and highly complex cloud environments. Honeycomb
customers stop wasting precious time on engineering mysteries because they can
quickly solve them and know exactly how to create fast, reliable, and great
customer experiences.

In order to use this provider, you must have a Honeycomb account. You can get
started today with a [free account](http://ui.honeycomb.io/signup?&utm_source=terraform&utm_medium=partner&utm_campaign=signup&utm_keyword=&utm_content=free-product-signup).

Use the navigation to the left to read about the available resources and data
sources.

## Example usage

```hcl
terraform {
  required_providers {
    honeycombio = {
      source  = "honeycombio/honeycombio"
      version = "~> 0.37.0"
    }
  }
}

# Configure the Honeycomb provider
provider "honeycombio" {
  # You can set the API key with the environment variable HONEYCOMB_API_KEY,
  # or the HONEYCOMB_KEY_ID+HONEYCOMB_KEY_SECRET environment variable pair
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

More advanced examples can be found in the [example
directory](https://github.com/honeycombio/terraform-provider-honeycombio/tree/main/example).

## A note on "Datasets"

Several resources in this provider accept a `dataset` or `datasets` argument to
specify which Honeycomb Dataset the resource belongs to. These resources include
but aren't limited to:

* queries
* triggers
* slos
* markers
* columns
* boards

Whenever a resource accepts a `dataset` or `datasets` argument, the argument is
expected to be a Dataset **slug**, not a Dataset name or ID. Dataset slugs can be
found in the URL of the dataset in the Honeycomb UI, or in the `slug` field of
the [Dataset API](https://api-docs.honeycomb.io/api/datasets/createdataset#datasets/createdataset/t=response&c=200&path=slug).

### Configuring the provider for Honeycomb EU

If you are a Honeycomb EU customer, to use the provider you must override the
default API host. This can be done with a `provider` block or by setting the
`HONEYCOMB_API_ENDPOINT` environment variable.

```hcl
provider "honeycombio" {
  api_url = "https://api.eu1.honeycomb.io"
}
```

## Authentication

The Honeycomb provider requires an API key to communicate with the Honeycomb APIs.
The provider can make calls to v1 and v2 APIs and requires specific key
configurations for each. For more information about API Keys, check out [Best
Practices for API Keys](https://docs.honeycomb.io/get-started/best-practices/api-keys/).

A single instance of the provider can be configured with both key types. At least
one of the v1 or v2 API key configuration is required.

### v1 APIs

v1 APIs require Configuration Keys. Their permissions can be managed in
_Environment settings_. Most resources and data sources call v1 APIs today.

The key can be set with the `api_key` argument or via the `HONEYCOMB_API_KEY` or
`HONEYCOMBIO_APIKEY` environment variable.

`HONEYCOMB_API_KEY` environment variable will take priority over the
`HONEYCOMBIO_APIKEY` environment variable.

### v2 APIs

v2 APIs require a Mangement Key. Their permissions can be managed in _Team
settings_. Resources and data sources that call v2 APIs will be noted along with
the scope required to use the resource or data source.

The key pair can be set with the `api_key_id` and `api_key_secret` arguments, or
via the `HONEYCOMB_KEY_ID` and `HONEYCOMB_KEY_SECRET` environment variables.

~> **Note** Hard-coding API keys in any Terraform configuration is not
recommended. Consider using the one of the environment variable options.

## Argument Reference

Arguments accepted by this provider include:

* `api_key` - (Optional) The Honeycomb API key to use. It can also be set using
  `HONEYCOMB_API_KEY` or `HONEYCOMBIO_APIKEY` environment variables.
* `api_key_id` - (Optional) The ID portion of the Honeycomb Management API key to
  use. It can also be set via the `HONEYCOMB_KEY_ID` environment variable.
* `api_key_secret` - (Optional) The secret portion of the Honeycomb Management
  API key to use. It can also be set via the `HONEYCOMB_KEY_SECRET` environment
  variable.
* `api_key_env_var` - (Optional) The name of the environment variable containing
  the Honeycomb API key. If not set, defaults to `HONEYCOMB_API_KEY`. Useful for
  multi-project scenarios where different projects need different API keys.
* `api_key_id_env_var` - (Optional) The name of the environment variable
  containing the Honeycomb API key ID. If not set, defaults to `HONEYCOMB_KEY_ID`.
  Useful for multi-project scenarios where different projects need different API
  keys.
* `api_key_secret_env_var` - (Optional) The name of the environment variable
  containing the Honeycomb API key secret. If not set, defaults to
  `HONEYCOMB_KEY_SECRET`. Useful for multi-project scenarios where different
  projects need different API keys.
* `api_url` - (Optional) Override the URL of the Honeycomb.io API. It can also be
  set using `HONEYCOMB_API_ENDPOINT`. Defaults to `https://api.honeycomb.io`.
* `debug` - (Optional) Enable to log additional debug information. To view the
  logs, set `TF_LOG` to at least debug.

At least one of `api_key`, or the `api_key_id` and `api_key_secret` pair must be
configured.

### Custom Environment Variable Configuration

To follow [Honeycomb's API Key Best
Practices](https://docs.honeycomb.io/get-started/best-practices/api-keys/), you
should use different API keys for different purposes. For example, the API key
used for production should be different from the one used for testing, and the key
used by your build process should be different from either of those.

This feature allows you to specify custom environment variable names to support
proper API key separation:

```hcl
provider "honeycombio" {
  api_key_env_var        = "HONEYCOMB_API_KEY_PROD"
  api_key_id_env_var     = "HONEYCOMB_KEY_ID_PROD"
  api_key_secret_env_var = "HONEYCOMB_KEY_SECRET_PROD"
}
```

This allows you to set different environment variables for different purposes:

* `HONEYCOMB_API_KEY_PROD` for production infrastructure management
* `HONEYCOMB_API_KEY_NONPROD` for non-production/testing infrastructure
* `HONEYCOMB_API_KEY_DEV` for development environments

This is particularly useful with CI/CD systems where you can set different
environment variables per project or environment. For example, you could:

1. **Set environment variables at the system level** for different projects
2. **Use different Terraform workspaces** with different environment variables
3. **Configure environment variables in your CI/CD pipeline** per project

Then in your respective project's `main.tf` files:

```hcl
# prod/main.tf
provider "honeycombio" {
  api_key_env_var        = "HONEYCOMB_API_KEY_PROD"
  api_key_id_env_var     = "HONEYCOMB_KEY_ID_PROD"
  api_key_secret_env_var = "HONEYCOMB_KEY_SECRET_PROD"
}

# nonprod/main.tf
provider "honeycombio" {
  api_key_env_var        = "HONEYCOMB_API_KEY_NONPROD"
  api_key_id_env_var     = "HONEYCOMB_KEY_ID_NONPROD"
  api_key_secret_env_var = "HONEYCOMB_KEY_SECRET_NONPROD"
}




For more detailed documentation and examples, see [Custom Environment Variables](custom_environment_variables.md).
