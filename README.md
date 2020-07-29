# Honeycomb.io Terraform Provider

![CI](https://github.com/kvrhdn/terraform-provider-honeycombio/workflows/CI/badge.svg)

_This is not an official Honeycomb.io provider!_

## Using the provider

_I'm trying to add this provider to the Terraform Registry, but for now you have to download and install the provider manually._

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

  # required, can also be set using environment variable HONEYCOMBIO_DATASET
  dataset = "my-dataset"
}
```

Examples of resources can be found in the [examples directory](example/). Documentation can be found in the [docs directory](docs/).

If you wish to manage multiple datasets, you can create multiple instances of the provider using aliases. Refer to the documentation [`alias`: Multiple Provider Instances](https://www.terraform.io/docs/configuration/providers.html#alias-multiple-provider-instances).

## License

This software is distributed under the terms of the MIT license, see [LICENSE](./LICENSE) for details.
