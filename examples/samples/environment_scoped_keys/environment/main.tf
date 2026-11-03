# The automation of a single Environment, which only has the environment-scoped
# Management Key made by the central configuration, and cannot reach the other Environments.
terraform {
  required_providers {
    honeycombio = {
      source = "honeycombio/honeycombio"
    }
  }
}

variable "key_id" {
  type = string
}

variable "key_secret" {
  type      = string
  sensitive = true
}

provider "honeycombio" {
  api_key_id     = var.key_id
  api_key_secret = var.key_secret
}

# With only a Key ID and Secret pair, the data source tells which Environment the key is for.
data "honeycombio_auth_metadata" "current" {}

output "environment" {
  value = data.honeycombio_auth_metadata.current.environment.slug
}

# Resources which use the v1 API, such as datasets, columns, queries and triggers,
# do not accept an environment-scoped key yet, and need a Configuration Key.
# provider "honeycombio" { api_key = var.configuration_key }
