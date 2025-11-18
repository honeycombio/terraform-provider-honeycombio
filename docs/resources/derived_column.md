# Resource: honeycombio_derived_column

Creates a derived column. For more information about derived columns, check out [Calculate with derived columns](https://docs.honeycomb.io/working-with-your-data/customizing-your-query/derived-columns/).

## Example Usage

### Dataset-specific

```hcl
variable "dataset" {
  type = string
}

resource "honeycombio_derived_column" "any_error" {
  alias       = "dc.any_error"
  expression  = "COALESCE($error.message, $app.legacy_error)"

  description = "Collapse OTEL semantic convention and legacy error messages into one field"
  dataset = var.dataset
}
```

### Environment-wide

```hcl
resource "honeycombio_derived_column" "duration_ms_log" {
  alias       = "dc.duration_ms_log10"
  expression  = "LOG10($duration_ms)"

  description = "LOG10 of duration_ms"
}
```

### Complex Formula

```hcl
resource "honeycombio_derived_column" "sli_calculation" {
  # Multi-line expression with HEREDOC syntax
  expression = <<DC
  IF(
    $service.name = "Backend" AND $name = "HandleRequest",
    !EXISTS($error.message)
  )
  DC

  alias       = "sli.errors"
  description = "Return true if any request error in the `Backend` service"
}
```

## Argument Reference

The following arguments are supported:

* `alias` - (Required) The name of the derived column. Must be unique per dataset.
* `expression` - (Required) The formula of the derived column. See [Derived Column Syntax](https://docs.honeycomb.io/reference/derived-column-formula/syntax/).
* `dataset` - (Optional) The dataset this derived column is added to. If not set, an Environment-wide derived column will be created.
* `description` - (Optional) A description that is shown in the UI.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the derived column.

## Import

Dataset-specific derived columns can be imported using a combination of the dataset name and their alias, e.g.

```
$ terraform import honeycombio_derived_column.my_column my-dataset/any_error
```

Environment-wide derived columns can be imported using just the alias, e.g.

```
$ terraform import honeycombio_derived_column.my_column duration_ms_log10
```
