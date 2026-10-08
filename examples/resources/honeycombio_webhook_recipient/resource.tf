# Standard Webhook Example

resource "honeycombio_webhook_recipient" "prod" {
  name   = "Production Alerts"
  secret = "a63dab148496ecbe04a1a802ca9b95b8"
  url    = "https://my.url.corp.net"
}

# Custom Webhook Example

resource "honeycombio_webhook_recipient" "prod" {
  name   = "Production Alerts"
  secret = "a63dab148496ecbe04a1a802ca9b95b8"
  url    = "https://my.url.corp.net"
    
  header {
    name = "Authorization"
    value = "Bearer 123"
  }
    
  template {
    type = "trigger"
    body = <<EOT
		{
			"name": " {{ .Name }}",
			"id": " {{ .ID }}",
			"description": " {{ .Description }}",
            "threshold": {
              "op": "{{ .Operator }}",
              "value": "{{ .Threshold }}"
            },
		}
		EOT
  }
    
  template {
    type = "anomaly"
    body = <<EOT
		{
			"service": "{{ .ServiceName }}",
			"signal": "{{ .SignalType }}",
			"status": "{{ .Alert.Status }}",
			"peak": {{ toJson .Anomaly.PeakValue }},
			"duration": {{ toJson .Anomaly.Duration }}
		}
		EOT
  }

  variable {
      name          = "severity"
      default_value = "critical"
  }
}
