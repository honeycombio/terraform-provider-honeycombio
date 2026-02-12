package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

var _ validator.String = triggerQuerySpecValidator{}

type triggerQuerySpecValidator struct{}

func (v triggerQuerySpecValidator) Description(_ context.Context) string {
	return "value must be a valid Trigger Query Specification"
}

func (v triggerQuerySpecValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v triggerQuerySpecValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	var q client.QuerySpec
	if err := json.Unmarshal([]byte(request.ConfigValue.ValueString()), &q); err != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			v.Description(ctx),
			fmt.Sprintf("%q: %s", request.ConfigValue.ValueString(), err.Error()),
		))
		return
	}

	// Reject HEATMAP and CONCURRENCY calculations
	for _, calc := range q.Calculations {
		if calc.Op == client.CalculationOpHeatmap {
			response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
				request.Path,
				v.Description(ctx),
				"Trigger queries cannot use HEATMAP calculations.",
			))
		}
		if calc.Op == client.CalculationOpConcurrency {
			response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
				request.Path,
				v.Description(ctx),
				"Trigger queries cannot use CONCURRENCY calculations.",
			))
		}
	}

	// Reject forbidden fields
	if q.Orders != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			v.Description(ctx),
			"Trigger queries cannot use orders.",
		))
	}
	if q.Limit != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			v.Description(ctx),
			"Trigger queries cannot use limit.",
		))
	}
	if q.StartTime != nil || q.EndTime != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			v.Description(ctx),
			"Trigger queries cannot use start_time or end_time.",
		))
	}
	if q.Granularity != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			v.Description(ctx),
			"Trigger queries cannot use granularity.",
		))
	}

	// Build list of calculations that don't match havings
	var calculationsWithoutHavings []client.CalculationSpec
	for _, calc := range q.Calculations {
		matchesHaving := false
		for _, having := range q.Havings {
			if reflect.DeepEqual(having.Column, calc.Column) &&
				having.CalculateOp != nil &&
				*having.CalculateOp == calc.Op {
				matchesHaving = true
				break
			}
		}
		if !matchesHaving {
			calculationsWithoutHavings = append(calculationsWithoutHavings, calc)
		}
	}

	// Enforce single non-having calculation
	var numCalculations int
	switch {
	case len(q.Calculations) == 1:
		numCalculations = 1
	case len(q.Havings) == 0:
		numCalculations = len(q.Calculations)
	default:
		numCalculations = len(calculationsWithoutHavings)
	}

	if numCalculations != 1 {
		var namesList []string
		for _, calc := range calculationsWithoutHavings {
			s := string(calc.Op)
			if calc.Column != nil {
				s = fmt.Sprintf("%s(%s)", s, *calc.Column)
			}
			namesList = append(namesList, s)
		}
		names := strings.Join(namesList, ", ")

		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			v.Description(ctx),
			fmt.Sprintf(
				"Trigger queries must contain a single calculation, but found %s",
				names,
			),
		))
	}
}

// ValidTriggerQuerySpec determines if the provided JSON is a valid Trigger Query Specification
func ValidTriggerQuerySpec() validator.String {
	return triggerQuerySpecValidator{}
}
