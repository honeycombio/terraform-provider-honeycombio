####################################################
# Example: Ensure required columns exist
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
    "db.system"   = "string",
    "db.type"     = "string",
    "duration_ms" = "float",
    "error"       = "boolean",
    "http.flavor" = "string",
  }
}

# Call the Columns API to ensure that the required columns exist
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
