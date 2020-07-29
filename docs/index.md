# Honeycomb.io Provider

The Honeycomb.io provider is used to interact with the Honeycomb.io API.

Use the navigation to the left to read about the available resources.

## Example usage

```hcl
# Configure the Honeycomb.io provider
provider "honeycombio" {
  # You can also set the environment variable HONEYCOMBIO_APIKEY
  api_key = "<your API key>"

  # You can also set the environment variable HONEYCOMBIO_DATASET
  dataset = "<your dataset>"
}

# Create a marker
resource "honeycombio_marker" "hello" {
    message = "Hello world!"
}
```

~> **Note**: Hard-coding API keys into any Terraform configuration is not recommended. Consider using the `HONEYCOMBIO_APIKEY` environment variable.

## Argument Reference

Arguments accepted by this provider include:

* `api_key` - (Required) The Honeycomb API key to use. It can also be set using the `HONEYCOMBIO_APIKEY` environment variable.
* `dataset` - (Required) The Honeycomb dataset to configure. It can also be set using the `HONEYCOMBIO_DATASET` environment variable.
