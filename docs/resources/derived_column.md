# Resource: honeycombio_derived_column

Creates a derived column. For more information about derived columns, check out [Calculate with derived columns](https://docs.honeycomb.io/working-with-your-data/customizing-your-query/derived-columns/).

## Example Usage

```hcl
variable "dataset" {
  type = string
}

resource "honeycombio_derived_column" "duration_ms_log" { 
  alias       = "duration_ms_log10"
  expression  = "LOG10($duration_ms)"
  description = "LOG10 of duration_ms"

  dataset = var.dataset
}

resource "honeycombio_derived_column" "example_multiline_expr" { 
  alias       = "example_multiline_expr"
  description = "Example of multiline expression"
 
  expression  = trimspace(<<EOT
  IF(
    REG_MATCH($request, `PATCH https://www.foo.[a-z]+:443/gw/api/[0-9]+/bar?`),
    AND(
      OR(
        STARTS_WITH($backend_status_code, "2"), 
        STARTS_WITH($backend_status_code, "3"), 
        STARTS_WITH($backend_status_code, "4")
      ),
      LTE($backend_processing_time, 2)
    )
  )
  EOT
  )

  dataset = var.dataset
}
```

## Argument Reference

The following arguments are supported:

* `dataset` - (Required) The dataset this derived column is added to.
* `alias` - (Required) The name of the derived column. Must be unique per dataset.
* `expression` - (Required) The function of the derived column. See [Derived Column Syntax](https://docs.honeycomb.io/working-with-your-data/customizing-your-query/derived-columns/#derived-column-syntax).
* `description` - (Optional) A description that is shown in the UI.

~> **Note** Use `trimline()` for multiline values of `expression`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the derived column.

## Import

Derived columns can be imported using a combination of the dataset name and their alias, e.g.

```
$ terraform import honeycombio_derived_column.my_column my-dataset/duration_ms_log10
```
