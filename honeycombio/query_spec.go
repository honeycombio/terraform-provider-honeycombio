package honeycombio

import (
	"encoding/json"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

type querySpecValidateDiagFunc func(q *honeycombio.QuerySpec) diag.Diagnostics

// validateQueryJSON checks that the input can be deserialized as a QuerySpec
// and optionally runs a list of custom validation functions.
func validateQueryJSON(validators ...querySpecValidateDiagFunc) schema.SchemaValidateDiagFunc {
	return func(i interface{}, path cty.Path) diag.Diagnostics {
		var q honeycombio.QuerySpec

		err := json.Unmarshal([]byte(i.(string)), &q)
		if err != nil {
			return diag.Errorf("value of query_json is not a valid query specification")
		}

		var diagnostics diag.Diagnostics

		for _, validator := range validators {
			diagnostics = append(diagnostics, validator(&q)...)
		}
		return diagnostics
	}
}
