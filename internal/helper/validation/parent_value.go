package validation

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ParentValueValidator struct {
	Expression path.Expression

	Value types.Dynamic
}

// Description returns a plaintext string describing the validator.
func (v ParentValueValidator) Description(_ context.Context) string {
	return fmt.Sprintf("If configured, be used only if %s is %s", v.Expression.String(), v.Value.String())
}

// MarkdownDescription returns a Markdown formatted string describing the validator.
func (v ParentValueValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ParentValueValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	// If the Value is unknown or null, there is nothing to validate.
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	// Find paths matching the Expression in the configuration data.
	matchedPaths, diags := req.Config.PathMatches(ctx, v.Expression)

	resp.Diagnostics.Append(diags...)

	// Collect all errors
	if diags.HasError() {
		return
	}

	// For each matched path, get the data and compare.
	for _, matchedPath := range matchedPaths {
		// Fetch the generic attr.Value at the given path. This ensures any
		// potential parent Value of a different type, which can be a null
		// or unknown Value, can be safely checked without raising a type
		// conversion error.
		var matchedPathValue attr.Value

		diags := req.Config.GetAttribute(ctx, matchedPath, &matchedPathValue)

		resp.Diagnostics.Append(diags...)

		// Collect all errors
		if diags.HasError() {
			continue
		}

		// If the matched path Value is null or unknown, we cannot compare
		// values, so continue to other matched paths.
		if matchedPathValue.IsNull() || matchedPathValue.IsUnknown() {
			continue
		}

		if !v.Value.UnderlyingValue().Equal(matchedPathValue) {
			resp.Diagnostics.AddAttributeError(
				matchedPath,
				"Invalid Attribute Value",
				fmt.Sprintf("%s must be equal to Value: %s", req.Path, v.Value.String()),
			)
		}
	}
}
