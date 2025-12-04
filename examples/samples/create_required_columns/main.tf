####################################################
# Example: Ensure required columns exist
#
# We call the Columns API directly to ensure that
# the required columns exist.
####################################################

variable "dataset" {
  type = string
}

variable "honeycomb_api_endpoint" {
  type    = string
  default = "https://api.honeycomb.io"
}

# A list of columns and their types
#  See: https://docs.honeycomb.io/api/tag/Columns
locals {
  required_columns = {
    # name           type
    "db.system"   = "string",
    "db.type"     = "string",
    "duration_ms" = "float",
    "error"       = "boolean",
    "http.flavor" = "string",
  }
}

# Ensure the required columns exist. Example for Terraform 1.3 and below
resource "null_resource" "ensure_columns" {
  for_each = local.required_columns

  provisioner "local-exec" {
    command = <<-EOT
curl -s ${var.honeycomb_api_endpoint}/1/columns/${var.dataset} \
  -X POST \
  -H "X-Honeycomb-Team: $${HONEYCOMB_API_KEY}" \
  -d '{"key_name": "${each.key}", "type": "${each.value}"}'
EOT
  }
}

# The same as above but with the `terraform_data` resource for Terraform 1.4+
resource "terraform_data" "ensure_columns" {
  for_each = local.required_columns

  provisioner "local-exec" {
    command = <<-EOT
curl -s ${var.honeycomb_api_endpoint}/1/columns/${var.dataset} \
  -X POST \
  -H "X-Honeycomb-Team: $${HONEYCOMB_API_KEY}" \
  -d '{"key_name": "${each.key}", "type": "${each.value}"}'
EOT
  }
}
