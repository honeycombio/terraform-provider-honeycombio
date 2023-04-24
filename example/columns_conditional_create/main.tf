####################################################
# Example: Create only the missing columns
####################################################

variable "dataset" {
  type = string
}

# returns all columns in a dataset
data "honeycombio_columns" "all" {
  dataset = var.dataset
}

# A list of columns and their types that are necessary
locals {
  required_columns = {
    "db.system"              = "string",
    "db.type"                = "string",
    "duration_ms"            = "float",
    "error"                  = "boolean",
    "http.flavor"            = "string",
  }

  cols_to_create = setsubtract(keys(local.required_columns), data.honeycombio_columns.all.names)
}

resource "honeycombio_column" "required_columns" {
  for_each = toset(local.cols_to_create)

  name    = each.key
  type    = lookup(local.required_columns, each.key, "string")
  dataset = var.dataset
}

output "columns_created" {
  value = local.cols_to_create
}
