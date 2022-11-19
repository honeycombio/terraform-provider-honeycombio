# Resource: honeycombio_dataset_definition

Dataset Definitions define the fields in your Dataset that have special meaning.

-> **Note** Some Dataset Definitions are automatically set when a dataset is created or first receives an event.

## Example Usage

```hcl
resource "honeycombio_dataset_definition" "trace-id" {
  dataset = var.dataset

  name   = "trace.trace_id"
  column = "trace_id"
}
```

## Argument Reference

The following arguments are supported:

- `dataset` - (Required) The dataset to set the Dataset Definition for.
- `name` - (Required) The name of the definition being set. See chart below for possible values.
- `column` - The column to set the definition to. Must be the name of an existing Column or the alias of an existing Derived Column.

### List of Dataset Definitions to be configured

Definition Name    | Description              
------------------ | -------------------------
`span_id`          | Span ID
`trace_id`         | Trace ID
`parent_id`        | Parent Span ID
`name`             | Name
`service_name`     | Service Name
`duration_ms`      | Span Duration
`span_kind`        | Metadata: Kind
`annotation_type`  | Metadata: Annotation Type
`link_span_id`     | Metadata: Link Span ID
`link_trace_id`    | Metadata: Link Trace ID
`error`            | Error
`status`           | HTTP Status Code
`route`            | Route
`user`             | User

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `column_type` - The type of the column the dataset definition is set to
