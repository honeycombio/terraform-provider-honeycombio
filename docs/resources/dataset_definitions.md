# Resource: honeycombio_dataset_definition

Dataset Definitions define the fields in your Dataset that have special meaning.

-> Some Dataset Definitions are automatically set when a dataset is created or first receives an event.

## Example Usage

```hcl
resource "honeycombio_dataset_definition" "trace_id" {
  dataset = var.dataset

  name   = "trace_id"
  column = "trace.trace_id"
}
```

## Argument Reference

The following arguments are supported:

- `dataset` - (Required) The dataset to set the Dataset Definition for.
- `name` - (Required) The name of the definition being set. See chart below for possible values.
- `column` - (Required) The column to set the definition to. Must be the name of an existing Column or the alias of an existing Derived Column.

### List of Dataset Definitions to be configured

Definition Name    | Description              
------------------ | -------------------------
`error`            | Error
`status`           | HTTP Status Code
`span_kind`        | Metadata: Kind
`annotation_type`  | Metadata: Annotation Type
`link_span_id`     | Metadata: Link Span ID
`link_trace_id`    | Metadata: Link Trace ID
`log_message`      | Log Event Message
`log_severity`     | Log Event Severity
`name`             | Name
`parent_id`        | Parent Span ID
`route`            | Route
`service_name`     | Service Name
`duration_ms`      | Span Duration
`span_id`          | Span ID
`trace_id`         | Trace ID
`user`             | User

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `column_type` - The type of the column assigned to the definition. Will be one of `column` or `derived_column`.

## Import

Dataset Definitions cannot be imported.
