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
	t.Parallel()

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
				{ID: types.StringValue("abcd12345")}, // we use ID over type+target when there is no state
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
			name: "remote notification details replace state details",
			args: args{
				remote: []client.NotificationRecipient{
					{
						ID:   "abcd12345",
						Type: client.RecipientTypeWebhook,
						Details: &client.NotificationRecipientDetails{
							Variables: []client.NotificationVariable{{Name: "severity", Value: "critical"}},
						},
					},
				},
				state: notificationRecipientModelsToSet([]models.NotificationRecipientModel{
					{
						ID:      types.StringValue("abcd12345"),
						Details: notificationRecipientDetailsWithVariablesToList(client.NotificationVariable{Name: "severity", Value: "warning"}),
					},
				}),
			},
			want: notificationRecipientModelsToSet([]models.NotificationRecipientModel{
				{
					ID:      types.StringValue("abcd12345"),
					Details: notificationRecipientDetailsWithVariablesToList(client.NotificationVariable{Name: "severity", Value: "critical"}),
				},
			}),
		},
		{
			name: "import includes remote notification details",
			args: args{
				remote: []client.NotificationRecipient{
					{
						ID:   "abcd12345",
						Type: client.RecipientTypeWebhook,
						Details: &client.NotificationRecipientDetails{
							Variables: []client.NotificationVariable{{Name: "severity", Value: "warning"}},
						},
					},
				},
				state: types.SetNull(elemType),
			},
			want: notificationRecipientModelsToSet([]models.NotificationRecipientModel{
				{
					ID:      types.StringValue("abcd12345"),
					Details: notificationRecipientDetailsWithVariablesToList(client.NotificationVariable{Name: "severity", Value: "warning"}),
				},
			}),
		},
		{
			name: "remote default pagerduty severity is ignored when unconfigured",
			args: args{
				remote: []client.NotificationRecipient{
					{
						ID:      "abcd12345",
						Type:    client.RecipientTypePagerDuty,
						Details: &client.NotificationRecipientDetails{PDSeverity: client.PDDefaultSeverity},
					},
				},
				state: notificationRecipientModelsToSet([]models.NotificationRecipientModel{
					{ID: types.StringValue("abcd12345")},
				}),
			},
			want: notificationRecipientModelsToSet([]models.NotificationRecipientModel{
				{ID: types.StringValue("abcd12345")},
			}),
		},
		{
			name: "remote non-default pagerduty severity is drift when unconfigured",
			args: args{
				remote: []client.NotificationRecipient{
					{
						ID:      "abcd12345",
						Type:    client.RecipientTypePagerDuty,
						Details: &client.NotificationRecipientDetails{PDSeverity: client.PDSeverityINFO},
					},
				},
				state: notificationRecipientModelsToSet([]models.NotificationRecipientModel{
					{ID: types.StringValue("abcd12345")},
				}),
			},
			want: notificationRecipientModelsToSet([]models.NotificationRecipientModel{
				{
					ID:      types.StringValue("abcd12345"),
					Details: types.ListValueMust(types.ObjectType{AttrTypes: models.NotificationRecipientDetailsAttrType}, notificationRecipientDetailsToValue("info")),
				},
			}),
		},
		{
			name: "configured default pagerduty severity is preserved",
			args: args{
				remote: []client.NotificationRecipient{
					{
						ID:      "abcd12345",
						Type:    client.RecipientTypePagerDuty,
						Details: &client.NotificationRecipientDetails{PDSeverity: client.PDDefaultSeverity},
					},
				},
				state: notificationRecipientModelsToSet([]models.NotificationRecipientModel{
					{
						ID:      types.StringValue("abcd12345"),
						Details: types.ListValueMust(types.ObjectType{AttrTypes: models.NotificationRecipientDetailsAttrType}, notificationRecipientDetailsToValue("critical")),
					},
				}),
			},
			want: notificationRecipientModelsToSet([]models.NotificationRecipientModel{
				{
					ID:      types.StringValue("abcd12345"),
					Details: types.ListValueMust(types.ObjectType{AttrTypes: models.NotificationRecipientDetailsAttrType}, notificationRecipientDetailsToValue("critical")),
				},
			}),
		},
		{
			name: "remote variables removed outside of terraform",
			args: args{
				remote: []client.NotificationRecipient{
					{ID: "abcd12345", Type: client.RecipientTypeWebhook},
				},
				state: notificationRecipientModelsToSet([]models.NotificationRecipientModel{
					{
						ID:      types.StringValue("abcd12345"),
						Details: notificationRecipientDetailsWithVariablesToList(client.NotificationVariable{Name: "severity", Value: "warning"}),
					},
				}),
			},
			want: notificationRecipientModelsToSet([]models.NotificationRecipientModel{
				{
					ID: types.StringValue("abcd12345"),
					Details: types.ListValueMust(types.ObjectType{AttrTypes: models.NotificationRecipientDetailsAttrType}, []attr.Value{
						types.ObjectValueMust(models.NotificationRecipientDetailsAttrType, map[string]attr.Value{
							"pagerduty_severity": types.StringNull(),
							"variable":           types.SetNull(types.ObjectType{AttrTypes: models.NotificationVariableAttrType}),
						}),
					}),
				},
			}),
		},
		{
			name: "import ignores default pagerduty severity",
			args: args{
				remote: []client.NotificationRecipient{
					{
						ID:      "abcd12345",
						Type:    client.RecipientTypePagerDuty,
						Details: &client.NotificationRecipientDetails{PDSeverity: client.PDDefaultSeverity},
					},
				},
				state: types.SetNull(elemType),
			},
			want: notificationRecipientModelsToSet([]models.NotificationRecipientModel{
				{ID: types.StringValue("abcd12345")},
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
					Details: types.ListValueMust(types.ObjectType{AttrTypes: models.NotificationRecipientDetailsAttrType}, notificationRecipientDetailsToValue("warning")),
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
					{ID: types.StringValue("ijkl13579"), Details: types.ListValueMust(types.ObjectType{AttrTypes: models.NotificationRecipientDetailsAttrType}, notificationRecipientDetailsToValue("warning"))},
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
					{ID: types.StringValue("ijkl13579"), Details: types.ListValueMust(types.ObjectType{AttrTypes: models.NotificationRecipientDetailsAttrType}, notificationRecipientDetailsToValue("warning"))},
				}),
			},
			want: notificationRecipientModelsToSet([]models.NotificationRecipientModel{
				{ID: types.StringValue("efgh67890"), Type: types.StringValue("slack"), Target: types.StringValue("#test-foo")},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			got := reconcileReadNotificationRecipientState(context.Background(), tt.args.remote, tt.args.state, &diags)

			assert.Empty(t, diags)
			assert.Equal(t, tt.want, got)
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

func notificationRecipientDetailsToValue(s string) []attr.Value {
	return []attr.Value{types.ObjectValueMust(models.NotificationRecipientDetailsAttrType, map[string]attr.Value{"pagerduty_severity": types.StringValue(s), "variable": types.SetNull(types.ObjectType{AttrTypes: models.NotificationVariableAttrType})})}
}

func notificationRecipientDetailsWithVariablesToList(variables ...client.NotificationVariable) types.List {
	variableValues := make([]attr.Value, 0, len(variables))
	for _, variable := range variables {
		variableValues = append(variableValues, types.ObjectValueMust(models.NotificationVariableAttrType, map[string]attr.Value{
			"name":  types.StringValue(variable.Name),
			"value": types.StringValue(variable.Value),
		}))
	}

	variableSet := types.SetValueMust(types.ObjectType{AttrTypes: models.NotificationVariableAttrType}, variableValues)
	details := types.ObjectValueMust(models.NotificationRecipientDetailsAttrType, map[string]attr.Value{
		"pagerduty_severity": types.StringNull(),
		"variable":           variableSet,
	})

	return types.ListValueMust(types.ObjectType{AttrTypes: models.NotificationRecipientDetailsAttrType}, []attr.Value{details})
}

func Test_expandNotificationRecipients(t *testing.T) {
	var diags diag.Diagnostics
	got := expandNotificationRecipients(context.Background(), notificationRecipientModelsToSet([]models.NotificationRecipientModel{
		{ID: types.StringValue("abcd12345")},
		{
			ID:      types.StringValue("efgh67890"),
			Details: notificationRecipientDetailsWithVariablesToList(client.NotificationVariable{Name: "severity", Value: "warning"}),
		},
	}), &diags)

	assert.Empty(t, diags)
	assert.ElementsMatch(t, []client.NotificationRecipient{
		// recipients without details send an empty object to clear any existing details
		{ID: "abcd12345", Details: &client.NotificationRecipientDetails{}},
		{
			ID: "efgh67890",
			Details: &client.NotificationRecipientDetails{
				Variables: []client.NotificationVariable{{Name: "severity", Value: "warning"}},
			},
		},
	}, got)
}
