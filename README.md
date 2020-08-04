# Honeycomb.io Terraform Provider

[![CI](https://github.com/kvrhdn/terraform-provider-honeycombio/workflows/CI/badge.svg)](https://github.com/kvrhdn/terraform-provider-honeycombio/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/kvrhdn/terraform-provider-honeycombio)](https://goreportcard.com/report/github.com/kvrhdn/terraform-provider-honeycombio)
[![codecov](https://codecov.io/gh/kvrhdn/terraform-provider-honeycombio/branch/main/graph/badge.svg)](https://codecov.io/gh/kvrhdn/terraform-provider-honeycombio)

_This is not an official Honeycomb.io provider!_

Want to contribute? Check out [CONTRIBUTING.md](./CONTRIBUTING.md).

Questions? Feel free to create a new issue or find us on the **Honeycomb Pollinators** Slack, channel **#terraform-provider** (you can find a link to request an invite [here](https://www.honeycomb.io/blog/spread-the-love-appreciating-our-pollinators-community/)).

## Using the provider

This provider is published on the [Terraform Registry](https://registry.terraform.io/providers/kvrhdn/honeycombio/latest)! You can try that out if you're using Terraform v0.13. If not, you'll have to install the provider manually.

### Building the provider

You need to have Go 1.14+ installed to build this provider.

First, clone this repository and build the executable:

```go
go build -o terraform-provider-honeycombio
```

Next you can either:
- place the executable in your working directory (the directory where you run `terraform init`).
- install is as described here: [Third-party Plugins](https://www.terraform.io/docs/configuration/providers.html#third-party-plugins).

### Adding the provider

Declare the provider in your Terraform configuration and run `terraform init`.

```hcl
provider "honeycombio" {
  # required, can also be set using environment variable HONEYCOMBIO_APIKEY
  api_key = "<your API key>"
}
```

Examples of resources can be found in the [examples directory](example/). Documentation can be found in the [docs directory](docs/).

## License

This software is distributed under the terms of the MIT license, see [LICENSE](./LICENSE) for details.
