# Data Source: honeycombio_dataset

The `honeycombio_dataset` data source retrieves the details of a single Dataset.
If you want to retrieve multiple Datasets, use the `honeycombio_datasets` data source instead.

-> **API Keys** Note that this requires a [v1 API Key](https://registry.terraform.io/providers/honeycombio/honeycombio/latest/docs#v1-apis)

## Example Usage

```hcl
# Retrieve the details of a Dataset
data "honeycombio_dataset" "my-service" {
  slug = "my-service"
}
```

## Argument Reference

The following arguments are supported:

* `slug` - (Required) The Slug of the Dataset

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `name` - the Dataset's name.
* `description` - the Dataset's description.
* `expand_json_depth` - The Dataset's maximum unpacking depth of nested JSON fields.
* `delete_protected` - the current state of the Dataset's deletion protection status.
* `created_at` - ISO8601-formatted time the dataset was created.
* `last_written_at` - ISO8601-formatted time the dataset was last written to (received event data).

