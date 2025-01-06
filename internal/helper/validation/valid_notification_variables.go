package validation

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

var _ validator.Set = notificationVariablesValidator{}

type notificationVariablesValidator struct {
	schemes []string
}

func (v notificationVariablesValidator) Description(_ context.Context) string {
	return "value must be a valid set of webhook variables"
}

func (v notificationVariablesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v notificationVariablesValidator) ValidateSet(ctx context.Context, request validator.SetRequest, response *validator.SetResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	// variable names cannot be duplicated
	var variables []models.NotificationVariableModel
	response.Diagnostics.Append(request.ConfigValue.ElementsAs(ctx, &variables, false)...)

	duplicateMap := make(map[string]bool)
	for i, v := range variables {
		name := v.Name.ValueString()
		if duplicateMap[name] {
			response.Diagnostics.AddAttributeError(
				path.Root("variable").AtListIndex(i).AtName("name"),
				"Conflicting configuration arguments",
				"cannot have more than one \"variable\" with the same \"name\"",
			)
		}
		duplicateMap[name] = true
	}
}

// ValidQuerySpec determines if the provided JSON is a valid Honeycomb Query Specification
func ValidNotificationVariables() validator.Set {
	return notificationVariablesValidator{}
}
