data "honeycombio_auth_metadata" "current" {}

output "team_name" {
  value = data.honeycombio_auth_metadata.current.team.name
}

output "environment_slug" {
  value = data.honeycombio_auth_metadata.current.environment.slug
} 

output "slo_management_access" {
  value = data.honeycombio_auth_metadata.current.api_key_access.slos
}
