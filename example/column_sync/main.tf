variable "dataset" {
  type = string
}

variable "otherdataset" {
  type = string
}

# Retrieve the details of a single column
data "honeycombio_column" "mycol" {
  dataset = var.dataset
  name    = "mycol"
}

# Ensure a column in another dataset remains in sync
resource "honeycombio_column" "othercol" {
  dataset     = var.otherdataset
  name        = data.honeycombio_column.mycol.name
  type        = data.honeycombio_column.mycol.type
  description = data.honeycombio_column.mycol.description
  hidden      = data.honeycombio_column.mycol.hidden

  # but only if its updated_at is greater than an arbitrary date
  count = timecmp(data.honeycombio_column.mycol.updated_at, "2023-04-20T00:00:00Z") == 1 ? 1 : 0
}
