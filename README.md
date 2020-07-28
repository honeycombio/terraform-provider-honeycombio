# Honeycomb.io Terraform Provider

_This is not an official Honeycomb.io provider!_

## Build from source

You'll need Go installed to build this provider.

```go
go build -o terraform-provider-honeycombio
```

Next you can either:
- place the executable in your working directory (the directory where you run `terraform init`)
- install is as described here: https://www.terraform.io/docs/configuration/providers.html#third-party-plugins

## Docs

### Provider `honeycombio`

```hcl
provider "honeycombio" {
  # required, can also be set using environment variable HONEYCOMBIO_APIKEY
  api_key = "<your API key>"

  # required, can also be set using environment variable HONEYCOMBIO_DATASET
  dataset = "my-dataset"
}
```

If you wish to manage multiple datasets, you can create multiple instances of the provider using aliases. See the documentation here: https://www.terraform.io/docs/configuration/providers.html#alias-multiple-provider-instances

### Resource `honeycombio_marker`

Create a marker in the dataset. The marker is set at the time the resource is created by Terraform and will not be destroyed.

_⚠️ This Terraform resource is atypical in the sense that destroying the resource does not delete the marker. This is intentional since you don't want to delete markers when you set a new one. It is not possible to remove or update a marker using this provider, instead it will always create a new marker._

```hcl
resource "honeycombio_marker" "marker" {
  # optional
  message = "Hello world!"

  # optional
  type    = "deploys"

  # optional
  url     = "https://www.honeycomb.io/"
}
```

This resource uses the markers API: https://docs.honeycomb.io/api/markers/

## License

This software is distributed under the terms of the MIT license, see [LICENSE](./LICENSE) for details.
