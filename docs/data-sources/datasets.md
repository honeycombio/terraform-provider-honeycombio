# Data Source: honeycombio_datasets

The datasets data source allows the datasets of an account to be retrieved.

## Example Usage

```hcl
# returns all datasets
data "honeycombio_datasets" "all" {
}

# only returns the datasets starting with 'foo_'
data "honeycombio_datasets" "foo" {
  starts_with = "foo_"
}

# create a resource for every dataset
resource "honeycombio_derived_column" "duration_ms_log10" {
    for_each = toset(data.honeycombio_datasets.all.names)

    alias       = "duration_ms_log10"
    expression  = "LOG10($duration_ms)"
    description = "LOG10 of duration_ms"

    dataset = each.key
}
```

## Argument Reference

The following arguments are supported:

* `starts_with` - (Optional) Only return datasets starting with the given value.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `names` - a list of all the dataset names.
* `slugs` - a list of all the dataset slugs.
