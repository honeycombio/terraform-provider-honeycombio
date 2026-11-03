# A central configuration creates an Environment for each team, their Ingest Keys,
# and an environment-scoped Management Key for each Environment.
#
# Run it with a team-scoped Management Key which has the scopes `environments:write`
# and `api-keys:write`, which is also what creating the environment-scoped keys needs.
# HONEYCOMB_KEY_ID and HONEYCOMB_KEY_SECRET can be used to set the key.
terraform {
  required_providers {
    honeycombio = {
      source = "honeycombio/honeycombio"
    }
  }
}

variable "environments" {
  type    = set(string)
  default = ["dev", "staging", "prod"]
}

resource "honeycombio_environment" "env" {
  for_each = var.environments

  name = each.key
}

resource "honeycombio_api_key" "ingest" {
  for_each = honeycombio_environment.env

  name           = "${each.key} ingest"
  type           = "ingest"
  environment_id = each.value.id

  permissions {
    create_datasets = true
  }
}

resource "honeycombio_api_key" "automation" {
  for_each = honeycombio_environment.env

  name           = "${each.key} automation"
  type           = "management"
  environment_id = each.value.id
  scopes         = ["datasets:read", "datasets:write"]
}

# The secrets are only available when the keys are created, and are kept in the Terraform state.
# Pass them to the environment's own automation with a secrets manager
# instead of with these outputs, and keep the state private.
output "ingest_keys" {
  sensitive = true
  value     = { for env, key in honeycombio_api_key.ingest : env => key.key }
}

output "automation_keys" {
  sensitive = true
  value = {
    for env, key in honeycombio_api_key.automation : env => {
      key_id     = key.id
      key_secret = key.secret
    }
  }
}
