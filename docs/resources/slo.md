# Resource: honeycombio_slo

Creates a service level objective (SLO). For more information about SLOs, check out [Set Service Level Objectives (SLOs)](https://docs.honeycomb.io/working-with-your-data/slos/).

## Example Usage
### Single Dataset SLO

```hcl
resource "honeycombio_derived_column" "request_latency_sli" {
  alias       = "sli.request_latency"
  description = "SLI: request latency less than 300ms"
  dataset     = var.dataset

  # heredoc also works
  expression = file("../sli/sli.request_latency.honeycomb")

  lifecycle {
    create_before_destroy = true
  }
}

resource "honeycombio_slo" "slo" {
  name              = "Latency SLO"
  description       = "example of an SLO"
  dataset           = var.dataset
  sli               = honeycombio_derived_column.request_latency_sli.alias
  target_percentage = 99.9
  time_period       = 30
}
```

### Multi-Dataset SLO

```hcl
resource "honeycombio_derived_column" "request_latency_sli" {
  alias       = "sli.request_latency"
  description = "SLI: request latency less than 300ms"
  dataset     = "__all__"

  # heredoc also works
  expression = file("../sli/sli.request_latency.honeycomb")

  lifecycle {
    create_before_destroy = true
  }
}

resource "honeycombio_slo" "slo" {
  name              = "Latency SLO"
  description       = "example of an SLO"
  datasets          = [var.dataset1, var.dataset2]
  sli               = honeycombio_derived_column.request_latency_sli.alias
  target_percentage = 99.9
  time_period       = 30
}
```

-> **Note** As [Derived Columns](derived_column.md) cannot be renamed or deleted while in use, it is recommended to use the [create_before_destroy](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle#create_before_destroy) lifecycle argument on your SLI resources as shown in the example above.
This way you will avoid running into conflicts if the Derived Column needs to be recreated.

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the SLO.
* `description` - (Optional) A description of the SLO's intent and context.
* `dataset` - (Optional) The dataset this SLO is created in. Must be the same dataset as the SLI unless the SLI's dataset is `"__all__"`. Conflicts with `datasets`. 
* `datasets` - (Optional) Array of datasets the SLO is evaluated on. Conflicts with `dataset`. Must have a length between 1 and 10.
* `sli` - (Required) The alias of the Derived Column that will be used as the SLI to indicate event success.
The derived column used as the SLI must be in the same dataset as the SLO. Additionally,
the column evaluation should consistently return nil, true, or false, as these are the only valid values for an SLI.
* `target_percentage` - (Required) The percentage of qualified events that you expect to succeed during the `time_period`.
* `time_period` - (Required) The time period, in days, over which your SLO will be evaluated.

~> **Note** `dataset` will be deprecated in a future release. In the meantime, you can swap `dataset` with a single value array for `datasets` to effectively evaluate to the same configuration.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the SLO.

## Import

SLOs can be imported using a combination of the dataset name and their ID, e.g.

```
$ terraform import honeycombio_slo.my_slo my-dataset/bj9BwOb1uKz
```

You can find the ID in the URL bar when visiting the SLO from the UI.
