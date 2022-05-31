# Resource: honeycombio_slo

Creates a service level objective (SLO). For more information about SLOs, check out [Set Service Level Objectives (SLOs)](https://docs.honeycomb.io/working-with-your-data/slos/).

## Example Usage

```hcl
resource "honeycombio_derived_column" "request_latency_sli" {
  alias       = "sli.request_latency"
  description = "SLI: request latency less than 300ms"
  dataset     = var.dataset

  # heredoc also works
  expression = file("../sli/sli.request_latency.honeycomb")
}

resource "honeycombio_slo" "slo" {
  name              = "Latency SLO"
  description       = "example of an SLO
  dataset           = var.dataset
  sli               = honeycombio_derived_column.request_latency_sli.alias
  target_percentage = 99.9
  time_period       = 30
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the SLO.
* `description` - (Optional) A description of the SLO's intent and context.
* `dataset` - (Required) The dataset this SLO is created in. Must be the same dataset as the SLI.
* `sli` - (Required) The alias of the Derived Column that will be used as the SLI to indicate event success.
The derived column used as the SLI must be in the same dataset as the SLO. Additionally,
the column evaluation should consistently return nil, true, or false, as these are the only valid values for an SLI.
* `target_percentage` - (Required) The percentage of qualified events that you expect to succeed during the `time_period`.
* `time_period` - (Required) The time period, in days, over which your SLO will be evaluated.
* 
## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the SLO.
