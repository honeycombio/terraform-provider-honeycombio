variable "dataset" {
  type = string
}
resource "honeycombio_marker_setting" "deploy_marker" {
  type    =  "deploy"
  color   = "#DF4661"
  dataset = var.dataset
}
