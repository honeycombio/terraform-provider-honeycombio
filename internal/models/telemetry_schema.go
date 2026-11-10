package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TelemetrySchemaResourceModel struct {
	ID         types.String `tfsdk:"id"`
	SchemaYAML types.String `tfsdk:"schema_yaml"`
}
