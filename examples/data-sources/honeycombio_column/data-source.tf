variable "dataset" {
  type = string
}

# Retrieve the details of a single column
data "honeycombio_column" "mycol" {
  dataset = var.dataset
  name    = "mycol"
}
