# Honeycomb.io Terraform Provider

[![OSS Lifecycle](https://img.shields.io/osslifecycle/honeycombio/terraform-provider-honeycombio)](https://github.com/honeycombio/home/blob/main/honeycomb-oss-lifecycle-and-practices.md)
[![CI](https://github.com/honeycombio/terraform-provider-honeycombio/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/honeycombio/terraform-provider-honeycombio/actions/workflows/ci.yaml)
[![Terraform Registry](https://img.shields.io/github/v/release/honeycombio/terraform-provider-honeycombio?color=5e4fe3&label=Terraform%20Registry&logo=terraform&sort=semver)](https://registry.terraform.io/providers/honeycombio/honeycombio/latest)

A Terraform provider for Honeycomb.io.

📄 Check out [the documentation](https://registry.terraform.io/providers/honeycombio/honeycombio/latest/docs)

🏗️ Examples can be found in [example/](example/)

❓ Questions? Feel free to create a new issue or find us on the **Honeycomb Pollinators** Slack, channel [**#discuss-api-and-terraform**](https://honeycombpollinators.slack.com/archives/C017T9FFT0D) (you can find a link to request an invite [here](https://www.honeycomb.io/blog/spread-the-love-appreciating-our-pollinators-community/))

🔧 Want to contribute? Check out [CONTRIBUTING.md](./CONTRIBUTING.md)

## Using the provider

You can install the provider directly from the [Terraform Registry](https://registry.terraform.io/providers/honeycombio/honeycombio/latest).
Add the following block in your Terraform config, this will download the provider from the Terraform Registry:

```hcl
terraform {
  required_providers {
    honeycombio = {
      source  = "honeycombio/honeycombio"
      version = "~> 0.43.0"
    }
  }
}
```

You can override the default API Endpoint (`https://api.honeycomb.io`) by setting the `HONEYCOMB_API_ENDPOINT` environment variable.

The Honeycomb provider requires an API key to communicate with the Honeycomb APIs.
The provider can make calls to v1 and v2 APIs and requires specific key configurations for each.
For more information about API Keys, check out [Best Practices for API Keys](https://docs.honeycomb.io/get-started/best-practices/api-keys/).

A single instance of the provider can be configured with both key types.
At least one of the v1 or v2 API key configuration is required.

### v1 APIs

v1 APIs require Configuration Keys.
Their permissions can be managed in _Environment settings_.
Most resources and data sources call v1 APIs today.

The key can be set with the `api_key` argument or via the `HONEYCOMB_API_KEY` or `HONEYCOMBIO_APIKEY` environment variable.

`HONEYCOMB_API_KEY` environment variable will take priority over the `HONEYCOMBIO_APIKEY` environment variable.

### v2 APIs

v2 APIs require a Mangement Key.
Their permissions can be managed in _Team settings_.
Resources and data sources that call v2 APIs will be noted along with the scope required to use the resource or data source.

The key pair can be set with the `api_key_id` and `api_key_secret` arguments, or via the `HONEYCOMB_KEY_ID` and `HONEYCOMB_KEY_SECRET` environment variables.

### Configuring the provider for Honeycomb EU

If you are a Honeycomb EU customer, to use the provider you must override the default API host.
This can be done with a `provider` block (example below) or by setting the `HONEYCOMB_API_ENDPOINT` environment variable.

```hcl
provider "honeycombio" {
  api_url = "https://api.eu1.honeycomb.io"
}
```

## License

This software is distributed under the terms of the MIT license, see [LICENSE](./LICENSE) for details.
