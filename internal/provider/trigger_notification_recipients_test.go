package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

func Test_reconcileReadTriggerNotificationRecipientState(t *testing.T) {
	elemType := types.ObjectType{AttrTypes: models.TriggerNotificationRecipientAttrType}
	type args struct {
		remote []client.TriggerNotificationRecipient
		state  types.Set
	}
	tests := []struct {
		name string
		args args
		want types.Set
	}{
		{
			name: "both empty",
			args: args{},
			want: types.SetNull(elemType),
		},
		{
			// on import we cannot know how the recipient was authored, so the routing
			// attributes are nulled alongside type, target and notification_details
			name: "empty state nulls the routing attributes",
			args: args{
				remote: []client.TriggerNotificationRecipient{
					{
						NotificationRecipient: client.NotificationRecipient{
							ID: "abcd12345", Type: client.RecipientTypeEmail, Target: "test@example.com",
						},
						GroupFilter:         map[string][]string{"service.name": {"checkout"}},
						PDPerGroupIncidents: client.ToPtr(true),
					},
				},
				state: types.SetNull(elemType),
			},
			want: triggerNotificationRecipientModelsToSet([]models.TriggerNotificationRecipientModel{
				{
					NotificationRecipientModel: models.NotificationRecipientModel{
						ID: types.StringValue("abcd12345"),
					},
				},
			}),
		},
		{
			name: "a matched recipient keeps its stored routing",
			args: args{
				remote: []client.TriggerNotificationRecipient{
					{
						NotificationRecipient: client.NotificationRecipient{
							ID: "abcd12345", Type: client.RecipientTypePagerDuty, Target: "test-pagerduty",
						},
						GroupFilter:         map[string][]string{"service.name": {"checkout", "cart"}},
						PDPerGroupIncidents: client.ToPtr(true),
					},
				},
				state: triggerNotificationRecipientModelsToSet([]models.TriggerNotificationRecipientModel{
					{
						NotificationRecipientModel: models.NotificationRecipientModel{
							ID: types.StringValue("abcd12345"),
						},
						GroupFilter:         groupFilterValue(map[string][]string{"service.name": {"checkout", "cart"}}),
						PDPerGroupIncidents: types.BoolValue(true),
					},
				}),
			},
			want: triggerNotificationRecipientModelsToSet([]models.TriggerNotificationRecipientModel{
				{
					NotificationRecipientModel: models.NotificationRecipientModel{
						ID: types.StringValue("abcd12345"),
					},
					GroupFilter:         groupFilterValue(map[string][]string{"service.name": {"checkout", "cart"}}),
					PDPerGroupIncidents: types.BoolValue(true),
				},
			}),
		},
		{
			// an unmatched recipient is taken from the API, so its routing comes through
			name: "an unmatched recipient takes its routing from the API",
			args: args{
				remote: []client.TriggerNotificationRecipient{
					{
						NotificationRecipient: client.NotificationRecipient{
							ID: "efgh67890", Type: client.RecipientTypeSlack, Target: "#test-alerts",
						},
						GroupFilter: map[string][]string{"service.name": {"frontend"}},
					},
				},
				state: triggerNotificationRecipientModelsToSet([]models.TriggerNotificationRecipientModel{
					{
						NotificationRecipientModel: models.NotificationRecipientModel{
							ID: types.StringValue("abcd12345"),
						},
					},
				}),
			},
			want: triggerNotificationRecipientModelsToSet([]models.TriggerNotificationRecipientModel{
				{
					NotificationRecipientModel: models.NotificationRecipientModel{
						ID:     types.StringValue("efgh67890"),
						Type:   types.StringValue("slack"),
						Target: types.StringValue("#test-alerts"),
					},
					GroupFilter: groupFilterValue(map[string][]string{"service.name": {"frontend"}}),
				},
			}),
		},
		{
			// the API returns pagerduty_per_group_incidents=false for every PagerDuty
			// recipient, which must not become a diff against a config that omits it
			name: "an unmatched PagerDuty recipient normalizes a false per-group flag to null",
			args: args{
				remote: []client.TriggerNotificationRecipient{
					{
						NotificationRecipient: client.NotificationRecipient{
							ID: "ijkl13579", Type: client.RecipientTypePagerDuty, Target: "test-pagerduty",
						},
						PDPerGroupIncidents: client.ToPtr(false),
					},
				},
				state: triggerNotificationRecipientModelsToSet([]models.TriggerNotificationRecipientModel{
					{
						NotificationRecipientModel: models.NotificationRecipientModel{
							ID: types.StringValue("abcd12345"),
						},
					},
				}),
			},
			want: triggerNotificationRecipientModelsToSet([]models.TriggerNotificationRecipientModel{
				{
					NotificationRecipientModel: models.NotificationRecipientModel{
						ID:     types.StringValue("ijkl13579"),
						Type:   types.StringValue("pagerduty"),
						Target: types.StringValue("test-pagerduty"),
					},
					PDPerGroupIncidents: types.BoolNull(),
				},
			}),
		},
		{
			// a catch-all recipient has no group_filter at all
			name: "an absent group filter becomes null rather than an empty map",
			args: args{
				remote: []client.TriggerNotificationRecipient{
					{
						NotificationRecipient: client.NotificationRecipient{
							ID: "efgh67890", Type: client.RecipientTypeSlack, Target: "#test-oncall",
						},
						GroupFilter: map[string][]string{},
					},
				},
				state: triggerNotificationRecipientModelsToSet([]models.TriggerNotificationRecipientModel{
					{
						NotificationRecipientModel: models.NotificationRecipientModel{
							ID: types.StringValue("abcd12345"),
						},
					},
				}),
			},
			want: triggerNotificationRecipientModelsToSet([]models.TriggerNotificationRecipientModel{
				{
					NotificationRecipientModel: models.NotificationRecipientModel{
						ID:     types.StringValue("efgh67890"),
						Type:   types.StringValue("slack"),
						Target: types.StringValue("#test-oncall"),
					},
					GroupFilter: types.MapNull(models.GroupFilterValueType),
				},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := diag.Diagnostics{}
			got := reconcileReadTriggerNotificationRecipientState(context.Background(), tt.args.remote, tt.args.state, &diags)
			assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_flattenGroupFilter(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		filter map[string][]string
		want   types.Map
	}{
		"nil is null": {
			filter: nil,
			want:   types.MapNull(models.GroupFilterValueType),
		},
		"empty is null": {
			filter: map[string][]string{},
			want:   types.MapNull(models.GroupFilterValueType),
		},
		"populated round-trips": {
			filter: map[string][]string{"service.name": {"checkout"}},
			want:   groupFilterValue(map[string][]string{"service.name": {"checkout"}}),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := diag.Diagnostics{}
			got := flattenGroupFilter(context.Background(), tt.filter, &diags)
			assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_expandGroupFilter(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		filter types.Map
		want   map[string][]string
	}{
		"null is nil": {
			filter: types.MapNull(models.GroupFilterValueType),
			want:   nil,
		},
		"unknown is nil": {
			filter: types.MapUnknown(models.GroupFilterValueType),
			want:   nil,
		},
		// a zero-valued Map has a nil element type; it must still be treated as null
		"zero value is nil": {
			filter: types.Map{},
			want:   nil,
		},
		"populated round-trips": {
			filter: groupFilterValue(map[string][]string{"service.name": {"cart", "checkout"}}),
			want:   map[string][]string{"service.name": {"cart", "checkout"}},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := diag.Diagnostics{}
			got := expandGroupFilter(context.Background(), tt.filter, &diags)
			assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags.Errors())
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_flattenPerGroupIncidents(t *testing.T) {
	t.Parallel()

	assert.Equal(t, types.BoolNull(), flattenPerGroupIncidents(nil),
		"omitted by the API (a non-PagerDuty recipient) should be null")
	assert.Equal(t, types.BoolNull(), flattenPerGroupIncidents(client.ToPtr(false)),
		"false is indistinguishable from unset and should normalize to null")
	assert.Equal(t, types.BoolValue(true), flattenPerGroupIncidents(client.ToPtr(true)))
}

func triggerNotificationRecipientModelsToSet(n []models.TriggerNotificationRecipientModel) types.Set {
	var values []attr.Value
	for _, r := range n {
		values = append(values, triggerNotificationRecipientModelToObjectValue(context.Background(), r, &diag.Diagnostics{}))
	}
	return types.SetValueMust(types.ObjectType{AttrTypes: models.TriggerNotificationRecipientAttrType}, values)
}

func groupFilterValue(filter map[string][]string) types.Map {
	elements := make(map[string]attr.Value, len(filter))
	for column, values := range filter {
		set := make([]attr.Value, 0, len(values))
		for _, v := range values {
			set = append(set, types.StringValue(v))
		}
		elements[column] = types.SetValueMust(types.StringType, set)
	}
	return types.MapValueMust(models.GroupFilterValueType, elements)
}
