terraform {
  required_providers {
    honeycombio = {
      source  = "honeycombio/honeycombio"
      version = "~> 0.37.0"
    }
  }
}

# Configure the Honeycomb provider with custom environment variable names
# This allows different projects to use different API keys
provider "honeycombio" {
  # Use custom environment variable names for this project
  api_key_env_var        = "HONEYCOMB_API_KEY_PROD"
  api_key_id_env_var     = "HONEYCOMB_KEY_ID_PROD"
  api_key_secret_env_var = "HONEYCOMB_KEY_SECRET_PROD"
  
  # Optional: Override API URL for Honeycomb EU
  # api_url = "https://api.eu1.honeycomb.io"
}

# Example: Create a dataset for this environment
resource "honeycombio_dataset" "example" {
  name        = "terraform-example-dataset"
  description = "Example dataset created by Terraform"
}

# Example: Create a marker in the dataset
resource "honeycombio_marker" "deployment" {
  message = "Production deployment"
  dataset = honeycombio_dataset.example.name
}

# Example: Create a query
resource "honeycombio_query" "example" {
  dataset = honeycombio_dataset.example.name
  query_json = jsonencode({
    time_range = 3600
    granularity = 60
    breakdowns = ["app.team"]
    calculations = [{
      op = "COUNT"
    }]
  })
}

# Example: Create a trigger based on the query
resource "honeycombio_trigger" "example" {
  name    = "High Error Rate"
  dataset = honeycombio_dataset.example.name
  query_id = honeycombio_query.example.id
  
  threshold {
    op    = ">"
    value = 100
  }
  
  frequency = 300 # 5 minutes
} 