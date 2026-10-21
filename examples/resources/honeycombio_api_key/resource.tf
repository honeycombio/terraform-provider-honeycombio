resource "honeycombio_api_key" "prod_ingest" {
  name        = "Production Ingest"
  description = "Used by the checkout service in production"
  type        = "ingest"

  environment_id = var.environment_id

  permissions {
    create_datasets = true
  }
}

output "ingest_key" {
  value = "${honeycomb_api_key.prod_ingest.key}"
}
