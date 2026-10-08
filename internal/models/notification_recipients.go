package models

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NotificationRecipientModel struct {
	ID      types.String `tfsdk:"id"`
	Type    types.String `tfsdk:"type"`
	Target  types.String `tfsdk:"target"`
	Details types.List   `tfsdk:"notification_details"` // NotificationRecipientDetailsModel
}

var NotificationRecipientAttrType = map[string]attr.Type{
	"id":     types.StringType,
	"type":   types.StringType,
	"target": types.StringType,
	"notification_details": types.ListType{ElemType: types.ObjectType{
		AttrTypes: NotificationRecipientDetailsAttrType,
	}},
}

// TriggerNotificationRecipientModel is the `recipient` block of honeycombio_trigger.
// It extends NotificationRecipientModel with the per-group routing attributes, which are
// supported by Triggers only -- honeycombio_burn_alert keeps the base block.
//
// NotificationRecipientModel is embedded by value and carries no `tfsdk` tag of its own:
// the framework's reflection promotes its tagged fields into this object type. Embedding
// by pointer is not supported.
type TriggerNotificationRecipientModel struct {
	NotificationRecipientModel

	GroupFilter         types.Map  `tfsdk:"group_filter"` // map[string]Set of string
	PDPerGroupIncidents types.Bool `tfsdk:"pagerduty_per_group_incidents"`
}

// GroupFilterValueType is the element type of a recipient's `group_filter` -- a set of
// routing values per group by column.
var GroupFilterValueType = types.SetType{ElemType: types.StringType}

// TriggerNotificationRecipientAttrType is written out in full rather than derived from
// NotificationRecipientAttrType, which is a mutable package-level map still shared with
// honeycombio_burn_alert. Keep the two in sync by hand; the guard test in
// notification_recipients_test.go fails if an attribute is added to one and not the other.
var TriggerNotificationRecipientAttrType = map[string]attr.Type{
	"id":     types.StringType,
	"type":   types.StringType,
	"target": types.StringType,
	"notification_details": types.ListType{ElemType: types.ObjectType{
		AttrTypes: NotificationRecipientDetailsAttrType,
	}},
	"group_filter":                  types.MapType{ElemType: GroupFilterValueType},
	"pagerduty_per_group_incidents": types.BoolType,
}

type NotificationRecipientDetailsModel struct {
	PDSeverity types.String `tfsdk:"pagerduty_severity"`
	Variables  types.Set    `tfsdk:"variable"`
}

var NotificationRecipientDetailsAttrType = map[string]attr.Type{
	"pagerduty_severity": types.StringType,
	"variable":           types.SetType{ElemType: types.ObjectType{AttrTypes: NotificationVariableAttrType}},
}

type NotificationVariableModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

var NotificationVariableAttrType = map[string]attr.Type{
	"name":  types.StringType,
	"value": types.StringType,
}
