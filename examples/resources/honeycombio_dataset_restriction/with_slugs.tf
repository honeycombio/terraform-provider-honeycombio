# The Environment that holds the Dataset
data "honeycombio_environment" "prod" {
  detail_filter {
    name  = "slug"
    value = "prod"
  }
}

# A Dataset that Terraform does not manage
data "honeycombio_dataset" "billing" {
  slug = "billing"
}

resource "honeycombio_dataset_restriction" "billing" {
  environment_id = data.honeycombio_environment.prod.id
  dataset_id     = data.honeycombio_dataset.billing.hcid
}
