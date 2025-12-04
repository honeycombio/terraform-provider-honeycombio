resource "honeycombio_derived_column" "request_latency_sli" {
  alias       = "sli.request_latency"
  description = "SLI: request latency less than 300ms"

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

  tags = {
    team         = "red"
    experimental = "true"
  }
}
