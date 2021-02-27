# Contributing

All contributions are welcome, whether they are technical in nature or not.

Feel free to open a new issue to ask questions, discuss issues or propose enhancements. You can also chat with us on the **Honeycomb Pollinators** Slack in the **#terraform-provider** channel, you can find a direct link to request an invite in [Spread the Love: Appreciating Our Pollinators Community](https://www.honeycomb.io/blog/spread-the-love-appreciating-our-pollinators-community/).

The rest of this document describes how to get started developing on this repository.

## What should I know before I get started?

### Relevant documentation

Hashicorp has a lot of documentation on creating custom Terraform providers categorized under [Extending Terraform](https://www.terraform.io/docs/extend/index.html). This might help when getting started, but are not a pre-requisite to contribute. Feel free to just open an issue and we can guide you along the way.

We use [go-honeycombio](https://github.com/kvrhdn/go-honeycombio) to call the various Honeycomb APIs. While this takes care of most implementation details, it can still be valuable to check out [the official documentation of the APIs](https://docs.honeycomb.io/api/). The provider can only do as much as the APIs allow.

### What's in progress and what's next?

We maintain [an activity board](https://github.com/kvrhdn/terraform-provider-honeycombio/projects/1) with all the work that is currently being worked on and/or considered. Hopefully this can give a sense of what is next to come. The board is intended to create overview across the project, it's not a strict plan of action.

## Contributing changes

### Preview document changes

Hashicorp has a tool to preview documentation. Visit [registry.terraform.io/tools/doc-preview](https://registry.terraform.io/tools/doc-preview). 

### Running the test

Most of the tests are acceptance tests, which will call real APIs. To run the tests you'll need to have access to a Honeycomb account. If not, you can create a new free team.

First, **create an API key**. Initially you'll have to check all permissions, but _Send Events_ and _Create Datasets_ can be disabled once setup is done.

Next, **initialize the dataset by sending a test event**. This will 1) create the dataset and 2) create columns that are used in the tests.  We need the following columns to exist: `duration_ms`, `trace.parent_id` and `app.tenant`.

The easiest way to send an event is with [honeyvent](https://github.com/honeycombio/honeyvent):

```sh
# install honeyvent - you can also clone the repository and build it
go get github.com/honeycombio/honeyvent

# use honeyvent to send an event with dummy values
honeyvent -k <your API key> -d <dataset> -n duration_ms -v 100 -n trace.parent_id -v abc -n app.tenant -v def
```

Finally, **run the acceptance tests** by passing the API key and dataset as environment variables:

```sh
HONEYCOMBIO_APIKEY=<your API key> HONEYCOMBIO_DATASET=<dataset> make testacc
```

### Using a locally built version of the provider

It can be handy to run terraform with a local version of the honeycombio provider.

With **Terraform 0.12** this is very straightforward:

- build the provider with `go build`
- copy the executable (named `terraform-provider-honeycombio`) into your working directory
- run `terraform init`

This is a bit more involved with **Terraform 0.13**: the provider has to be installed in one of the [local mirror directories](https://www.terraform.io/docs/commands/cli-config.html#implied-local-mirror-directories) using the [new filesysem structure](https://www.terraform.io/upgrade-guides/0-13.html#new-filesystem-layout-for-local-copies-of-providers). Additinally, the provider should have a version.

For macOS, I've added the `install_macos` target in [`Makefile`](Makefile). Other OS's should be similar, feel free to add an additional target.

### Enabling log output

To print logs (including full dumps of requests and their responses), you have to set `TF_LOG` to at least `debug` and enable `HONEYCOMBIO_DEBUG` when running Terraform:

```sh
TF_LOG=debug HONEYCOMBIO_DEBUG=true terraform apply
```

A handy one-liner to simultaneously write the output to a file:

```sh
TF_LOG=debug HONEYCOMBIO_DEBUG=true terraform apply 2>&1 | tee output.log
```

For more information, see [Debugging Terraform](https://www.terraform.io/docs/internals/debugging.html).

### Style convention

CI will run the following tools to style code:

```sh
goimports -l -w .
go mod tidy
```

`goimports` will format the code like `gofmt` but will also fix imports. It can be installed with `go get golang.org/x/tools/cmd/goimports`.

Both commands should create no changes before a pull request can be merged.

### Run GitHub Actions in your fork

If you fork the repository, you can also run the tests on GitHub Actions (for free since it's a public repository). Unfortunatly there is no mechanism to share secrets, so all runs will fail until the necessary secrets are configured.

To properly setup the GitHub Actions, add the following secrets:

- `HONEYCOMBIO_APIKEY`: an API key for Honeycombio
- `HONEYCOMBIO_DATASET`: name of the test dataset
- `HONEYCOMBIO_DATASET_URL_ENCODED`: the same as `HONEYCOMBIO_DATASET`, but replace `/` with `-`. I.e. `foo/bar` becomes `foo-bar`.

## Release procedure

To release a new version of the Terraform provider a binary has to be built for a list of platforms ([more information](https://www.terraform.io/docs/registry/providers/publishing.html#creating-a-github-release)). This process is automated with GoReleaser and GitHub Actions.

- Create [a new release](https://github.com/kvrhdn/terraform-provider-honeycombio/releases/new)
- The tag and release title should be a semantic version
- To follow convention of other Terraform providers the description has the following sections (each section can be omitted if empty):
```
NOTES:
FEATURES:
ENHANCEMENTS:
BUG FIXES:
```
- After that tag has been created a GitHub Actions workflow Release will run and add binaries to the release (this workflow can run over 5 minutes)
- Once the tag is created, the [Terraform Registry](https://registry.terraform.io/providers/kvrhdn/honeycombio/latest) should also list the new version
