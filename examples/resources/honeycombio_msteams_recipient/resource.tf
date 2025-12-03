resource "honeycombio_msteams_recipient" "prod" {
  name = "Production Alerts"
  url  = "https://mycorp.webhook.office.com/webhookb2/abcd12345"
}
