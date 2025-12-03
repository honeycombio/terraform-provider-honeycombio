resource "honeycombio_pagerduty_recipient" "prod-oncall" {
  integration_key  = "cd6e8de3c857aefc950e0d5ebcb79ac2"
  integration_name = "Production on-call notifications"
}
