# A Management Key which is scoped to a single Environment.
resource "honeycombio_api_key" "prod_automation" {
  name           = "Production automation"
  description    = "Used by the production deployment pipeline"
  type           = "management"
  environment_id = var.environment_id

  # the provider's own key must hold every scope it grants
  scopes = ["datasets:read", "datasets:write"]
}

# The Key ID and Secret are only available when the key is created:
# store them in a secrets manager right away.
output "prod_automation_key_id" {
  value = honeycombio_api_key.prod_automation.id
}

output "prod_automation_key_secret" {
  value     = honeycombio_api_key.prod_automation.secret
  sensitive = true
}
