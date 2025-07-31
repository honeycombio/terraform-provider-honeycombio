# Custom Environment Variable Configuration Example

This example demonstrates how to configure the Honeycomb Terraform provider to use
custom environment variable names to follow [Honeycomb's API Key Best
Practices](https://docs.honeycomb.io/get-started/best-practices/api-keys/).

## Problem

According to Honeycomb's best practices, you should **"Use different API keys for
different purposes"**. For example:

- The API key used for production infrastructure management should be different
  from the one used for testing
- The key used by your build process should be different from either of those
- The key used for development environments should be different from production

The default provider configuration only supports hardcoded environment variable
names like `HONEYCOMB_API_KEY`, which makes it difficult to follow these best
practices and properly separate API keys by purpose.

### Atlantis-Specific Problem

This is particularly problematic with **Atlantis** deployments. Without custom
environment variable support, Atlantis would require a custom server-side workflow
to dynamically set `HONEYCOMB_API_KEY` based on the value in a file named
`.HONEYCOMB_API_KEY` for each project. This approach is complex and requires
maintaining custom server-side logic.

## Solution

This example shows how to use the new `api_key_env_var`, `api_key_id_env_var`,
and `api_key_secret_env_var` provider configuration options to specify custom
environment variable names, enabling you to follow Honeycomb's best practices for
API key separation.

### Atlantis Solution

With custom environment variables, you can simply:

1. Store different API keys in different environment variables (e.g.,
   `HONEYCOMB_API_KEY_PROD`, `HONEYCOMB_API_KEY_NONPROD`)
2. Configure each project's `main.tf` to use the appropriate environment variable
3. Set the environment variables at the system level or in your CI/CD pipeline

This eliminates the need for complex server-side workflows and provides a clean,
maintainable solution for Atlantis deployments.

## Usage

> **Note:** This example requires a newer version of the `honeycombio` provider
> that supports custom environment variable names. The linter may show errors for
> the new attributes until the provider is updated.

1. Set your custom environment variables:

   ```bash
   export HONEYCOMB_API_KEY_PROD="your-production-api-key"
   export HONEYCOMB_KEY_ID_PROD="your-production-key-id"
   export HONEYCOMB_KEY_SECRET_PROD="your-production-key-secret"
   ```

2. Run Terraform:

   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

## CI/CD Integration

For CI/CD systems, you can configure different environment variables per project
or environment. For example, you could:

1. **Set environment variables at the system level** for different projects
2. **Use different Terraform workspaces** with different environment variables
3. **Configure environment variables in your CI/CD pipeline** per project

### Atlantis Integration

This is particularly useful with **Atlantis** deployments. Instead of requiring
custom server-side workflows to dynamically set `HONEYCOMB_API_KEY` based on
project-specific files, you can:

1. Set environment variables at the system level (e.g., `HONEYCOMB_API_KEY_PROD`,
   `HONEYCOMB_API_KEY_NONPROD`)
2. Configure each project's `main.tf` to use the appropriate environment variable
3. Atlantis will automatically use the correct API key for each project without
   any custom server-side logic

Then in your respective project's `main.tf` files:

```hcl
# prod/main.tf
provider "honeycombio" {
  api_key_env_var        = "HONEYCOMB_API_KEY_PROD"
  api_key_id_env_var     = "HONEYCOMB_KEY_ID_PROD"
  api_key_secret_env_var = "HONEYCOMB_KEY_SECRET_PROD"
}

# nonprod/main.tf
provider "honeycombio" {
  api_key_env_var        = "HONEYCOMB_API_KEY_NONPROD"
  api_key_id_env_var     = "HONEYCOMB_KEY_ID_NONPROD"
  api_key_secret_env_var = "HONEYCOMB_KEY_SECRET_NONPROD"
}
```

## Benefits

- **Best Practices Compliance**: Follows Honeycomb's recommended approach of using
  different API keys for different purposes
- **Security**: API keys are scoped to specific purposes, minimizing the impact
  if one key is compromised
- **Flexibility**: No need to modify code when switching between environments or
  purposes
- **Purpose Separation**: Clear separation between production, testing, build,
  and development API keys
- **CI/CD Compatibility**: Works seamlessly with CI/CD systems for multi-project
  setups
