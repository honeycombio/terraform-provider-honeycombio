resource "honeycombio_derived_column" "duration_ms_log" {
  alias       = "dc.duration_ms_log10"
  expression  = "LOG10($duration_ms)"

  description = "LOG10 of duration_ms"
}
