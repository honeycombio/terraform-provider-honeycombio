# Honeycomb SLI Example

Honeycomb SLO is built on a Derived Column based-SLI which must evaluate to `true`, `false`, or `null`.

The SLI contains the bulk of the configuration as the qualification of an event by the SLI can become a bit complex.
Terraform can easily codify and manage the lifecycle of the Derived Column which makes up your SLI.

If you are a VSCode user consider using the [VSCode Extension for Honeycomb](https://marketplace.visualstudio.com/items?itemName=michaelcsickles.honeycomb-derived) to help the authoring experience.

For additional SLI examples, take a look at the [Honeycomb Resource Samples](https://github.com/honeycombio/honeycomb-resource-samples) repo.
