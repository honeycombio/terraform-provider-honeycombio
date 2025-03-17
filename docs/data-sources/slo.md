# Data Source: honeycombio_slo

The `honeycombio_slo` data source retrieves the details of a single SLO.  
If you want to retreive multiple SLOs, use the `honeycombio_slos` data source instead.

## Example Usage

```hcl
# Retrieve the details of a single SLO
data "honeycombio_slo" "myslo" {
  id      = "fS4WfA82ACt"
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Required) The ID of the SLO
* `dataset` - (Deprecated) No longer required. The dataset this SLO is associated with.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `name` - the name of the SLO.
* `description` - the SLO's description.
* `sli` - the alias of the Derived Column used as the SLO's SLI.
* `datasets` - A list of dataset slugs the SLO is evaluated on.
* `target_percentage` - the percentage of qualified events expected to succeed during the `time_period`.
* `time_period` - The time period, in days, over which the SLO is evaluated.