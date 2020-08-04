# Contributing

All contributions are welcome, whether they are technical in nature or not. Feel free to open a new issue to ask questions, discuss issues or propose enhancements. You can also chat with us on the **Honeycomb Pollinators** Slack in the **#terraform-provider** channel, you can find a direct link to request an invite in [Spread the Love: Appreciating Our Pollinators Community](https://www.honeycomb.io/blog/spread-the-love-appreciating-our-pollinators-community/).

The rest of this document describes how to get started developing on this repository.

## What should I know before I get started?

Hashicorp has a lot of documentation on creating custom Terraform providers categorized under [Extending Terraform](https://www.terraform.io/docs/extend/index.html). These might help out with getting started, but are not a pre-requisite to contribute. Feel free to just open an issue and we can guide you along the way.

We use [go-honeycombio](https://github.com/kvrhdn/go-honeycombio) to call the various Honeycomb APIs. While this takes care of most implementation details, it can still be interesting to check out [the documentation on the APIs here](https://docs.honeycomb.io/api/). The provider can only do as much as the APIs allow.

## Contributing technical changes

### Running the test

Most of the tests are accpetance tests, which will need to call real APIs. To run the tests you'll need to have access to a Honeycomb account. If not, you can create a new free team.

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

### Style convention

CI will run the following tools to style code:

```sh
goimports -l -w .
go mod tidy
```

`goimports` will format the code like `gofmt` but will also fix imports. It can be installed with `go get golang.org/x/tools/cmd/goimports`.

Both commands should create no changes for a pull request to get merged.

### Run GitHub Actions in your fork

If you fork the repository, you can also run the tests on GitHub Actions (without costs since it's a public repository). Unfortunatly there is no mechanism to share secrets, so all runs will fail until the necessary secrets are configured.

To properly setup the GitHub Actions, add the following secrets:

- `HONEYCOMBIO_APIKEY`: an API key for Honeycombio
- `HONEYCOMBIO_DATASET`: name of the test dataset
- `HONEYCOMBIO_DATASET_URL_ENCODED`: the same as `HONEYCOMBIO_DATASET`, but replace `/` with `-`. I.e. `foo/bar` becomes `foo-bar`.
