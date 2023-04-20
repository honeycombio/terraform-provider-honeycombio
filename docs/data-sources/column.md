# Data Source: honeycombio_column

The column data source retrieves the details of a single column

## Example Usage

```hcl
variable "dataset" {
  type = string
}

# Retrieve the details of a single column
data "honeycombio_column" "mycol" {
  dataset = var.dataset
  name    = "mycol"
}

# Ensure a column in another dataset remains in sync
resource "honeycombio_column" "othercol" {
  dataset     = "otherdataset"
  name        = data.honeycombio_column.mycol.name
  type        = data.honeycombio_column.mycol.type
  description = data.honeycombio_column.mycol.description
  hidden      = data.honeycombio_column.mycol.hidden

  # but only if its updated_at is greater than an arbitrary date
  count = timecmp(data.honeycombio_column.mycol.updated_at, "2023-04-20T00:00:00Z") == 1 ? 1 : 0
}

```

## Argument Reference

The following arguments are supported:

* `dataset` - (Required) The dataset this column is associated with
* `name` - (Required) The name of the column

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `name` - the name of the column
* `type` - the type of the column (string, integer, float, or boolean)
* `description` - the description of the column
* `hidden` - whether or not the column is hidden from the query builder and results
* `last_written` - the timestamp (string) the last time the column received data
* `created_at` - the timestamp (string) when the column was created
* `updated_at` - the timestamp (string) when the column's metadata (type, description, etc) was last changed
