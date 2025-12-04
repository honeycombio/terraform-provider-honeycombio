resource "honeycombio_derived_column" "sli_calculation" {
  # Multi-line expression with HEREDOC syntax
  expression = <<DC
  IF(
    $service.name = "Backend" AND $name = "HandleRequest",
    !EXISTS($error.message)
  )
  DC

  alias       = "sli.errors"
  description = "Return true if the request succeeded without error in the `Backend` service"
}
