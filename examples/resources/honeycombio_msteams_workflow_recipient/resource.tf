resource "honeycombio_msteams_workflow_recipient" "prod" {
  name = "Production Alerts"
  url  = "https://mycorp.westus.logic.azure.com/workflows/123456"
}
