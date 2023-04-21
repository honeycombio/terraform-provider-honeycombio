# Data Source: honeycombio_columns

The columns data source allows the columns of a dataset to be retrieved.

## Example Usage

```hcl
variable "dataset" {
  type = string
}

# returns all columns
data "honeycombio_columns" "all" {
  dataset = var.dataset
}

# only returns the columns starting with 'foo_'
data "honeycombio_columns" "foo" {
  dataset     = var.dataset
  starts_with = "foo_"
}

# Iterate through the list of retrieved columns
data "honeycombio_column" "all" {
  for_each = toset(data.honeycombio_columns.all.names)

  dataset = var.dataset
  name    = each.key
}

# And you can now use this to read data.honeycombio_column.all["mycolumn"].type

####################################################
# Example: Create only the missing columns
####################################################
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
```

## Argument Reference

The following arguments are supported:

* `dataset` - (Required) The dataset to retrieve the columns list from
* `starts_with` - (Optional) Only return columns starting with the given value.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `names` - a list of all the column names, use with toset()
