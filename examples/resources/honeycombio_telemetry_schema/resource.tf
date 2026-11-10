resource "honeycombio_telemetry_schema" "example" {
  # Read the schema from a file so that it is sent exactly as written.
  schema_yaml = file("${path.module}/telemetry-schema.yaml")
}
