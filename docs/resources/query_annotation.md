# Resource: honeycombio_query_annotation

Creates a query annotation in a dataset.

-> **API Keys** Note that this requires a [v1 API Key](https://registry.terraform.io/providers/honeycombio/honeycombio/latest/docs#v1-apis)

-> A query annotation points to a specific query. Any change to the query will result in a new query ID and the annotation will no longer apply.
If you use the "honeycombio_query_specification" to determine the `query_id` parameter (as in the example below), Terraform will destroy the old query annotation and create a new one.
If this is wrong for your use case, please open an issue in [honeycombio/terraform-provider-honeycombio](https://github.com/honeycombio/terraform-provider-honeycombio).

## Example Usage

```hcl
variable "dataset" {
  type = string
}

data "honeycombio_query_specification" "test_query" {
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }

  filter {
    column = "duration_ms"
    op     = ">"
    value  = 10
  }
}

resource "honeycombio_query" "test_query" {
  dataset    = var.dataset
  query_json = data.honeycombio_query_specification.test_query.json
}

resource "honeycombio_query_annotation" "test_annotation" {
	dataset     = var.dataset
	query_id    = honeycombio_query.test_query.id
	name        = "My Cool Query"
	description = "Describes my cool query (optional)"
}
```

## Argument Reference

The following arguments are supported:

* `dataset` - (Optional) The dataset this query annotation is added to. If not set, an Environment-wide query annotation will be created.
* `query_id` - (Required) The ID of the query that the annotation will be created on. Note that a query can have more than one annotation.
* `name` - (Required) The name to display as the query annotation.
* `description` - (Optional) The description to display as the query annotation. Defaults to an empty string.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the query annotation. Useful for adding it to a board.

## Import

Query Annotations can be imported using a combination of the dataset name and their ID, e.g.

```
$ terraform import honeycombio_query_annotation.my_query_annotation my-dataset/JL0Xp8SH0Dg
```

