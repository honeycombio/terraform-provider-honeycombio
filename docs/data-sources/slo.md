# Data Source: honeycombio_slo

The `honeycombio_slo` data source retrieves the details of a single SLO for a dataset.
If you want to retreive multiple SLOs, use the `honeycombio_slos` data source instead.

## Example Usage

```hcl
variable "dataset" {
  type = string
}

# Retrieve the details of a single SLO
data "honeycombio_slo" "myslo" {
  dataset = var.dataset
  id      = "fS4WfA82ACt"
}
```

## Argument Reference

The following arguments are supported:

* `dataset` - (Required) The dataset this SLO is associated with
* `id` - (Required) The ID of the SLO

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `name` - the name of the SLO.
* `description` - the SLO's description.
* `sli` - the alias of the Derived Column used as the SLO's SLI.
* `target_percentage` - the percentage of qualified events expected to succeed during the `time_period`.
* `time_period` - The time period, in days, over which the SLO is evaluated.
