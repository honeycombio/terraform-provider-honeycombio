package models_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// TestNotificationRecipientAttrTypes_staySynchronized guards the fork between the
// `recipient` block shared by honeycombio_burn_alert and the extended one used by
// honeycombio_trigger.
//
// TriggerNotificationRecipientAttrType is written out literally, so it does not track
// additions to NotificationRecipientAttrType. An attribute added to the shared block but
// missed here would not be a compile error -- it surfaces at runtime as an ElementsAs
// failure on the Trigger read path only. This test turns that into a build failure.
func TestNotificationRecipientAttrTypes_staySynchronized(t *testing.T) {
	t.Parallel()

	shared := models.NotificationRecipientAttrType
	trigger := models.TriggerNotificationRecipientAttrType

	// The attributes the Trigger block adds on top of the shared one.
	triggerOnly := map[string]bool{
		"group_filter":                  true,
		"pagerduty_per_group_incidents": true,
	}

	require.Len(t, shared, 4,
		"the shared recipient block changed shape; mirror the change in TriggerNotificationRecipientAttrType and update this test")
	require.Len(t, trigger, len(shared)+len(triggerOnly),
		"TriggerNotificationRecipientAttrType should be the shared block plus exactly the per-group routing attributes")

	for name, typ := range shared {
		actual, ok := trigger[name]
		require.True(t, ok, "%q is in the shared recipient block but missing from the Trigger block", name)
		assert.Equal(t, typ, actual, "%q has diverged between the shared and Trigger recipient blocks", name)
	}

	for name := range trigger {
		if triggerOnly[name] {
			continue
		}
		assert.Contains(t, shared, name,
			"%q is in the Trigger recipient block but is neither shared nor a known Trigger-only attribute", name)
	}
}

// TestTriggerNotificationRecipientModel_roundTrips asserts the model's `tfsdk` tags --
// including those promoted from the embedded NotificationRecipientModel -- line up exactly
// with the object type, in both directions. The framework requires a 1:1 match, and a
// mismatch is a runtime diagnostic rather than a compile error.
func TestTriggerNotificationRecipientModel_roundTrips(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := models.TriggerNotificationRecipientModel{
		NotificationRecipientModel: models.NotificationRecipientModel{
			ID:     types.StringValue("abcd1234"),
			Type:   types.StringNull(),
			Target: types.StringNull(),
			Details: types.ListNull(types.ObjectType{
				AttrTypes: models.NotificationRecipientDetailsAttrType,
			}),
		},
		GroupFilter: types.MapValueMust(models.GroupFilterValueType, map[string]attr.Value{
			"service.name": types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("checkout"),
			}),
		}),
		PDPerGroupIncidents: types.BoolValue(true),
	}

	obj, diags := types.ObjectValueFrom(ctx, models.TriggerNotificationRecipientAttrType, model)
	require.False(t, diags.HasError(),
		"model -> object failed, the tfsdk tags do not match the object type: %v", diags.Errors())

	var got models.TriggerNotificationRecipientModel
	diags = obj.As(ctx, &got, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError(), "object -> model failed: %v", diags.Errors())

	assert.Equal(t, model, got)
}
