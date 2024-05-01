# Ensure Required Columns Exist

Honeycomb requires that a column exists before it can be referenced by a query or derived column.

The `honeycombio_column` resource is helpful if or what you want to manage a column's lifecycle.
But, with that lifecycle management comes the requirement to import columns which already exist due to a naming conflict as is the norm with Terraform.

If you are simply looking to ensure that the columns used by your query or derived column exist -- maybe in a module to be used across various environments -- consider using this pattern with either the `terraform_data` or `null_resource` resources to handle column creation.

Note: the example assumes that your Honeycomb API key is available via the `HONEYCOMB_API_KEY` environment variable.
