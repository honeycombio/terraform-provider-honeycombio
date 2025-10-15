# Data Source: honeycombio_derived_column

The `honeycombio_derived_column` data source retrieves the details of a single derived column.

-> **API Keys** Note that this requires a [v1 API Key](https://registry.terraform.io/providers/honeycombio/honeycombio/latest/docs#v1-apis)

~> **Warning** Terraform will fail unless a derived column is returned by the search. Ensure that your search is specific enough to return a derived column.
If you want to match multiple derived columns, use the `honeycombio_derived_columns` data source instead.

## Example Usage

```hcl
variable "dataset" {
  type = string
}

# Retrieve the details of a single derived column
data "honeycombio_derived_column" "mydc" {
  dataset = var.dataset
  alias   = "mydc"
}
```

## Argument Reference

The following arguments are supported:

* `alias` - (Required) The alias of the column
* `dataset` - (Optional) The dataset this derived column is associated with. If not set, an Environment-wide lookup will be performed.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - the ID of the derived column.
* `description` - the description of the derived column
* `expression` - the expression of the derived column
