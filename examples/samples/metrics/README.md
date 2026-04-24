# Honeycomb Metrics Example

End-to-end sample showing how to use metrics-only calculation operators in
Terraform-managed Honeycomb resources:

- `COUNT_DATAPOINTS` — counts the number of datapoints reported by a metrics
  dataset. The `column` is optional; omit it for a total or set it to scope
  to a specific metric.
- `HISTOGRAM_COUNT` — counts events recorded in a histogram-typed metric
  column. The `column` is required.

The example wires the same query specifications into both alerting (triggers)
and visualization (a flexible board), so you can see how the operators flow
through each consumer.

## Requirements

- A Honeycomb metrics dataset with the metrics referenced in `main.tf`
  (`app.cumulative`, `app.histogram`). You can seed an example metrics
  dataset with `scripts/setup-metrics`.

## Usage

```sh
terraform init
terraform plan -var dataset=<your-metrics-dataset>
```
