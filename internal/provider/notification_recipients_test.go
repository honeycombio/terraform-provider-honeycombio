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

func Test_reconcileReadNotificationRecipientState(t *testing.T) {
	elemType := types.ObjectType{AttrTypes: models.NotificationRecipientAttrType}
	type args struct {
		remote []client.NotificationRecipient
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
			name: "empty state",
			args: args{
				remote: []client.NotificationRecipient{
					{ID: "abcd12345", Type: client.RecipientTypeEmail, Target: "test@example.com"},
				},
				state: types.SetNull(elemType),
			},
			want: notificationRecipientModelsToSet([]models.NotificationRecipientModel{
				{ID: types.StringValue("abcd12345"), Type: types.StringValue("email"), Target: types.StringValue("test@example.com")},
			}),
		},
		{
			name: "empty remote",
			args: args{
				remote: []client.NotificationRecipient{},
				state: notificationRecipientModelsToSet([]models.NotificationRecipientModel{
					{ID: types.StringValue("abcd12345"), Type: types.StringValue("email"), Target: types.StringValue("test@example.com")},
				}),
			},
			want: types.SetNull(elemType),
		},
		{
			name: "remote and state reconciled",
			args: args{
				remote: []client.NotificationRecipient{
					{ID: "abcd12345", Type: client.RecipientTypeEmail, Target: "test@example.com"},
					{ID: "efgh67890", Type: client.RecipientTypeSlack, Target: "#test-channel"},
				},
				state: notificationRecipientModelsToSet([]models.NotificationRecipientModel{
					{ID: types.StringValue("abcd12345")},                                           // defined by ID
					{Type: types.StringValue("slack"), Target: types.StringValue("#test-channel")}, // defined by type+target
				}),
			},
			want: notificationRecipientModelsToSet([]models.NotificationRecipientModel{
				{ID: types.StringValue("abcd12345")},
				{Type: types.StringValue("slack"), Target: types.StringValue("#test-channel")},
			}),
		},
		{
			name: "remote has additional recipients",
			args: args{
				remote: []client.NotificationRecipient{
					{ID: "abcd12345", Type: client.RecipientTypeEmail, Target: "test@example.com"},
					{ID: "efgh67890", Type: client.RecipientTypeSlack, Target: "#test-channel"},
					{ID: "qrsty3847", Type: client.RecipientTypeSlack, Target: "#test-alerts"},
					{
						ID:     "ijkl13579",
						Type:   client.RecipientTypePagerDuty,
						Target: "test-pagerduty",
						Details: &client.NotificationRecipientDetails{
							PDSeverity: client.PDSeverityWARNING,
						}},
				},
				state: notificationRecipientModelsToSet([]models.NotificationRecipientModel{
					{ID: types.StringValue("abcd12345")},                                           // defined by ID
					{Type: types.StringValue("slack"), Target: types.StringValue("#test-channel")}, // defined by type+target
				}),
			},
			want: notificationRecipientModelsToSet([]models.NotificationRecipientModel{
				{ID: types.StringValue("abcd12345")},
				{Type: types.StringValue("slack"), Target: types.StringValue("#test-channel")},
				{ID: types.StringValue("qrsty3847"), Type: types.StringValue("slack"), Target: types.StringValue("#test-alerts")},
				{
					ID:      types.StringValue("ijkl13579"),
					Type:    types.StringValue("pagerduty"),
					Target:  types.StringValue("test-pagerduty"),
					Details: types.ListValueMust(types.ObjectType{AttrTypes: models.NotificationRecipientDetailsAttrType}, severityStringToValue("warning")),
				},
			}),
		},
		{
			name: "state has additional recipients",
			args: args{
				remote: []client.NotificationRecipient{
					{ID: "efgh67890", Type: client.RecipientTypeSlack, Target: "#test-foo"},
				},
				state: notificationRecipientModelsToSet([]models.NotificationRecipientModel{
					{ID: types.StringValue("abcd12345")},
					{Type: types.StringValue("slack"), Target: types.StringValue("#test-foo")},
					{ID: types.StringValue("ijkl13579"), Details: types.ListValueMust(types.ObjectType{AttrTypes: models.NotificationRecipientDetailsAttrType}, severityStringToValue("warning"))},
				}),
			},
			want: notificationRecipientModelsToSet([]models.NotificationRecipientModel{
				{Type: types.StringValue("slack"), Target: types.StringValue("#test-foo")},
			}),
		},
		{
			name: "state has totally unmatched recipients",
			args: args{
				remote: []client.NotificationRecipient{
					{ID: "efgh67890", Type: client.RecipientTypeSlack, Target: "#test-foo"},
				},
				state: notificationRecipientModelsToSet([]models.NotificationRecipientModel{
					{ID: types.StringValue("abcd12345")},
					{Type: types.StringValue("slack"), Target: types.StringValue("#test-channel")},
					{ID: types.StringValue("ijkl13579"), Details: types.ListValueMust(types.ObjectType{AttrTypes: models.NotificationRecipientDetailsAttrType}, severityStringToValue("warning"))},
				}),
			},
			want: notificationRecipientModelsToSet([]models.NotificationRecipientModel{
				{ID: types.StringValue("efgh67890"), Type: types.StringValue("slack"), Target: types.StringValue("#test-foo")},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, reconcileReadNotificationRecipientState(context.Background(), tt.args.remote, tt.args.state, &diag.Diagnostics{}))
		})
	}
}

func notificationRecipientModelsToSet(n []models.NotificationRecipientModel) types.Set {
	var values []attr.Value
	for _, r := range n {
		values = append(values, notificationRecipientModelToObjectValue(context.Background(), r, &diag.Diagnostics{}))
	}
	return types.SetValueMust(types.ObjectType{AttrTypes: models.NotificationRecipientAttrType}, values)
}

func severityStringToValue(s string) []attr.Value {
	return []attr.Value{types.ObjectValueMust(models.NotificationRecipientDetailsAttrType, map[string]attr.Value{"pagerduty_severity": types.StringValue(s)})}
}
