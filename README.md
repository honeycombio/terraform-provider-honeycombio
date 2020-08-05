# Honeycomb.io Terraform Provider

[![CI](https://github.com/kvrhdn/terraform-provider-honeycombio/workflows/CI/badge.svg)](https://github.com/kvrhdn/terraform-provider-honeycombio/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/kvrhdn/terraform-provider-honeycombio)](https://goreportcard.com/report/github.com/kvrhdn/terraform-provider-honeycombio)
[![codecov](https://codecov.io/gh/kvrhdn/terraform-provider-honeycombio/branch/main/graph/badge.svg)](https://codecov.io/gh/kvrhdn/terraform-provider-honeycombio)
[![Terraform Registry](https://img.shields.io/github/v/release/kvrhdn/terraform-provider-honeycombio?color=5e4fe3&label=Terraform%20Registry&logo=terraform&sort=semver)](https://registry.terraform.io/providers/kvrhdn/honeycombio/latest)

A Terraform provider for Honeycomb.io.

📄 Check out [the documentation](https://registry.terraform.io/providers/kvrhdn/honeycombio/latest/docs)  
🏗️ Examples can be found in [example/](example/)  
❓ Questions? Feel free to create a new issue or find us on the **Honeycomb Pollinators** Slack, channel **#terraform-provider** (you can find a link to request an invite [here](https://www.honeycomb.io/blog/spread-the-love-appreciating-our-pollinators-community/))  
🔧 Want to contribute? Check out [CONTRIBUTING.md](./CONTRIBUTING.md)  

## Using the provider

If you are using Terraform 0.13, you can install the provider directly from the [Terraform Registry](https://registry.terraform.io/providers/kvrhdn/honeycombio/latest). To use the provider with Terraform 0.12 you'll have to install the provider manually.

### Terraform 0.13 (currently in beta)

Add the following block in your Terraform config. For more information, refer to [Automatic installation of third-party providers](https://github.com/hashicorp/terraform/tree/guide-v0.13-beta/provider-sources#terraform-v013-beta-automatic-installation-of-third-party-providers).

This will download the provider from the Terraform Registry:

```hcl
terraform {
  required_providers {
    honeycombio = {
      source  = "kvrhdn/honeycombio"
      version = "~> 0.0.2"
    }
  }
}
```

### Terraform 0.12

To use this provider with Terraform 0.12, you will need to download and install the executable yourself. You can download the latest version from the [releases page](https://github.com/kvrhdn/terraform-provider-honeycombio/releases) or build it directly from source:

```sh
# clone the repository and run:
go build -o terraform-provider-honeycombio
```

Once you have the executable (it should be named `terraform-provider-honeycombio`), you can either place it in your working directory (the directory you run `terraform init`) or [install it is a third-party plugin](https://www.terraform.io/docs/configuration/providers.html#third-party-plugins).

## License

This software is distributed under the terms of the MIT license, see [LICENSE](./LICENSE) for details.
