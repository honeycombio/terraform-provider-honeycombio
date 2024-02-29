# Honeycomb.io Terraform Provider

[![OSS Lifecycle](https://img.shields.io/osslifecycle/honeycombio/terraform-provider-honeycombio)](https://github.com/honeycombio/home/blob/main/honeycomb-oss-lifecycle-and-practices.md)
[![CI](https://github.com/honeycombio/terraform-provider-honeycombio/workflows/CI/badge.svg)](https://github.com/honeycombio/terraform-provider-honeycombio/actions)
[![codecov](https://codecov.io/gh/honeycombio/terraform-provider-honeycombio/branch/main/graph/badge.svg)](https://codecov.io/gh/honeycombio/terraform-provider-honeycombio)
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
      version = "~> 0.22.0"
    }
  }
}
```

Set the API key used by Terraform setting the `HONEYCOMB_API_KEY` environment variable.
You can override the default API Endpoint (`https://api.honeycomb.io`) by setting the `HONEYCOMB_API_ENDPOINT` environment variable.

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
