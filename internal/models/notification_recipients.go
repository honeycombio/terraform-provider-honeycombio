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
