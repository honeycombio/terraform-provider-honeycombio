# Contributing

All contributions are welcome, whether they are technical in nature or not.

Feel free to open a new issue to ask questions, discuss bugs, or propose enhancements.

You can also chat with us on the **Honeycomb Pollinators** Slack in the **#discuss-api-and-terraform** channel.

The rest of this document describes how to get started developing on this repository.

## What should I know before I get started?

### Relevant documentation

Hashicorp has a lot of documentation on creating Terraform providers categorized under [Plugin Development](https://developer.hashicorp.com/terraform/plugin).
This might help when getting started, but are not a pre-requisite to contribute.
Feel free to just open an issue and we can guide you along the way.

## Contributing changes

As there are currently no Honeycomb SDKs, this repository contains an embedded API Client for the [Honeycomb API](https://docs.honeycomb.io/api)
in the `client/` directory in the root of the repository.

The provider currently has a mix of the older [Plugin SDKv2](https://developer.hashicorp.com/terraform/plugin/sdkv2)
and newer [Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) resources and data sources.
Any new additions should be built with the Plugin Framework.

 * The Plugin SDK-based code is contained in the `honeycombio/` directory in the root of the repository.
 * The Plugin Framework-based code is contained in the `internal/provider` directory.

Any PRs reimplementing Plugin SDKv2 resources or datasources in the Plugin Framework will be enthusiastically accepted. 🙏

### Preview document changes

Hashicorp has a tool to preview documentation.
Visit [registry.terraform.io/tools/doc-preview](https://registry.terraform.io/tools/doc-preview).

### Generating documentation

Provider documentation is generated using [tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs).

To regenerate the documentation after making changes:

```sh
make docs
```

To verify documentation is up to date (used in CI):

```sh
make docs-check
```

#### Documentation structure

- `templates/` - Markdown templates (`.md.tmpl`) that control documentation structure
- `examples/` - Standalone `.tf` files referenced in templates via `{{tffile "path/to/example.tf"}}`
- `docs/` - Generated documentation (do not edit directly)

#### Adding or updating documentation

1. **Schema descriptions**: Update `Description` or `MarkdownDescription` fields in the Go schema definitions. Use `MarkdownDescription` when you need links or formatting.

2. **Examples**: Add or edit `.tf` files in `examples/resources/<resource_name>/` or `examples/data-sources/<data_source_name>/`. Reference them in templates using `{{tffile}}`.

3. **Custom content**: Edit the relevant `.md.tmpl` template in `templates/`. The `{{ .SchemaMarkdown | trimspace }}` directive inserts the auto-generated schema documentation.

4. **Regenerate**: Run `make docs` and commit both template and generated changes.

### Running the tests

Most of the tests are live integration tests against the [Honeycomb API](https://docs.honeycomb.io/api).
To run the tests you'll need to have access to a Honeycomb account and both a [Configuration Key](https://docs.honeycomb.io/get-started/best-practices/api-keys/#configuration-keys)
and a [Management Key](https://docs.honeycomb.io/get-started/best-practices/api-keys/#management-keys) with all permissions and scopes granted.

Some tests, such as those for SLOs and those for the Query Data API require access to a Pro or Enterprise team.

Additionally, tests for Slack recipients requires that the Slack authorization be [set up with the team ahead of time](https://docs.honeycomb.io/working-with-your-data/triggers/#slack)

Next, some of the embedded client tests require that you **initialize the dataset**.
The helper script [setup-testsuite-dataset](scripts/setup-testsuite-dataset) will create the dataset and required columns that are used in the tests. 
You will need to use Bash 4+ to run this script.

```sh
HONEYCOMB_API_KEY=<your API key> HONEYCOMB_DATASET=<dataset> ./scripts/setup-testsuite-dataset
```

Finally, **run the full testsuite**!
There is a `.env.example` template file checked into the root of the repository. Create a local `.env` file based on the template which can be used to store the relevant environment variables required for the tests:

- `HONEYCOMB_API_KEY`: a Configuration Key for a Honeycomb Team
- `HONEYCOMB_DATASET`: name of the test dataset to run tests against
- `HONEYCOMB_KEY_ID` and `HONEYCOMB_KEY_SECRET`: the v2 Management API Key pair for a Honeycomb Team

Or alternatively, you can set them directly and run the `testacc` make target:

```sh
HONEYCOMB_API_KEY=<CONFIGURATION KEY> HONEYCOMB_KEY_ID=<MGMT KEY ID> HONEYCOMB_KEY_SECRET=<MGMT KEY SECRET> HONEYCOMB_DATASET=<dataset> make testacc
```

### Using a locally built version of the provider

It can be handy to run terraform with a local version of the provider during development.

The best way to do this is with a [Development Override](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers).
There is already a `.terraformrc.local` file checked into the root of the repository which may be a helpful starting point.

### Enabling log output

To print logs (including full dumps of requests and their responses), you have to set `TF_LOG` to at least `debug` when running Terraform:

A handy one-liner to simultaneously write the output to a file:

```sh
TF_LOG=debug terraform apply 2>&1 | tee output.log
```

For more information, see [Debugging Terraform](https://developer.hashicorp.com/terraform/internals/debugging).

### Lints and Style

This project uses `golangci-lint` with the configuration at `.golangci.yml` in the root of the repository.

### Run GitHub Actions on your fork

If you fork the repository, you can also run the tests on GitHub Actions (for free since it's a public repository). Unfortunately there is no mechanism to share secrets, so all runs will fail until the necessary secrets are configured.

To properly setup the GitHub Actions, add the following secrets:

- `HONEYCOMB_API_KEY`: a Configuration Key for a Honeycomb Team
- `HONEYCOMB_DATASET`: name of the test dataset to run tests against
- `HONEYCOMB_KEY_ID` and `HONEYCOMB_KEY_SECRET`: the v2 Management API Key pair for a Honeycomb Team

## Release procedure

To release a new version of the Terraform provider a binary has to be built for a list of platforms ([more information](https://developer.hashicorp.com/terraform/registry/providers/publishing#creating-a-github-release)).
This process is largely automated with GoReleaser and GitHub Actions.

First, we have to create a "release commit" which updates `CHANGELOG.md` in the root of the repository.
- Update `CHANGELOG.md` in the root of the repository with the changes since the previous release (skipping those marked as `no-changelog`)
- If it's a non-patch release, also update references to the build version in the various examples (`README.md`, `docs/index.md`).
- Open the PR with these changes, and add the `no-changelog` tag to the PR so the change PR is not included in the generated release notes.
- Merge once approved

Once the release commit has landed on the `main` branch:

- Create a tag following semantic convention prefixed with a `v` (i.e. `v0.83.0`)
  - this will start the "release" workflow which builds various versions of the provider for target platforms and architectures.
    This can take up to 10 minutes to complete.
- When the release workflow completes, go to [releases](https://github.com/honeycombio/terraform-provider-honeycombio/releases/) and you'll find a draft release with the build artifacts attached.
- Update the release name to match the release tag (i.e. `v0.83.0`)
- Copy the section from `CHANGELOG.md` for this release into the release description.
- Publish the release.
- Within a few minutes, the [Terraform Registry](https://registry.terraform.io/providers/honeycombio/honeycombio/latest) should have picked up and published the new version.
- After the release has been published, update the latest version in the Pollinators api-and-terraform channel description.
