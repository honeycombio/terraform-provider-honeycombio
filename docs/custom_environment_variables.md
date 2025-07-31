# Custom Environment Variables for API Key Best Practices

## Overview

The Terraform provider for Honeycomb.io supports custom environment variable names
to help you follow [Honeycomb's API Key Best
Practices](https://docs.honeycomb.io/get-started/best-practices/api-keys/), which
recommend **"Use different API keys for different purposes"**.

By default, the provider uses these environment variable names:

- `HONEYCOMB_API_KEY` (v1 API key)
- `HONEYCOMBIO_APIKEY` (legacy v1 API key)
- `HONEYCOMB_KEY_ID` + `HONEYCOMB_KEY_SECRET` (v2 API key pair)

This feature allows you to specify custom environment variable names to support
proper API key separation by purpose.

## Problem Solved

This feature specifically solves a common problem with **Atlantis** deployments.
Without custom environment variable support, Atlantis would require a custom
server-side workflow to dynamically set `HONEYCOMB_API_KEY` based on the value in
a file named `.HONEYCOMB_API_KEY` for each project. This approach is complex and
requires maintaining custom server-side logic.

With custom environment variables, you can simply:

1. Store different API keys in different environment variables (e.g.,
   `HONEYCOMB_API_KEY_PROD`, `HONEYCOMB_API_KEY_NONPROD`)
2. Configure each project's `main.tf` to use the appropriate environment variable
3. Set the environment variables at the system level or in your CI/CD pipeline

This eliminates the need for complex server-side workflows and provides a clean,
maintainable solution.

## Configuration Options

The provider supports three optional configuration options for custom environment
variable names:

- `api_key_env_var` - Custom environment variable name for v1 API key
- `api_key_id_env_var` - Custom environment variable name for v2 API key ID
- `api_key_secret_env_var` - Custom environment variable name for v2 API key secret

## Usage

### Basic Example

```hcl
provider "honeycombio" {
  api_key_env_var        = "HONEYCOMB_API_KEY_PROD"
  api_key_id_env_var     = "HONEYCOMB_KEY_ID_PROD"
  api_key_secret_env_var = "HONEYCOMB_KEY_SECRET_PROD"
}
```

### Multi-Environment Setup

For different environments, you can use different environment variable names:

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

Then set the corresponding environment variables:

```bash
# For production
export HONEYCOMB_API_KEY_PROD="your-production-api-key"
export HONEYCOMB_KEY_ID_PROD="your-production-key-id"
export HONEYCOMB_KEY_SECRET_PROD="your-production-key-secret"

# For non-production
export HONEYCOMB_API_KEY_NONPROD="your-nonprod-api-key"
export HONEYCOMB_KEY_ID_NONPROD="your-nonprod-key-id"
export HONEYCOMB_KEY_SECRET_NONPROD="your-nonprod-key-secret"
```

## Benefits

- **Best Practices Compliance**: Follows Honeycomb's recommended approach of using
  different API keys for different purposes
- **Security**: API keys are scoped to specific purposes, minimizing the impact
  if one key is compromised
- **Flexibility**: No need to modify code when switching between environments or
  purposes
- **Backward Compatibility**: Existing configurations continue to work without
  modification

## Example

See the [Custom Environment Variable Configuration
Example](../example/multi_project_configuration/) for a complete working example.
