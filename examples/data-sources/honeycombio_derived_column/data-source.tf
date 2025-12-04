variable "dataset" {
  type = string
}

# Retrieve the details of a single derived column
data "honeycombio_derived_column" "mydc" {
  dataset = var.dataset
  alias   = "mydc"
}
