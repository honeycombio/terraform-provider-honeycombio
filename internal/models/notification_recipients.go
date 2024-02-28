package models

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NotificationRecipientModel struct {
	ID      types.String `tfsdk:"id"`
	Type    types.String `tfsdk:"type"`
	Target  types.String `tfsdk:"target"`
	Details types.List   `tfsdk:"notification_details"`
}

type NotificationRecipientDetailsModel struct {
	PDSeverity types.String `tfsdk:"pagerduty_severity"`
}

var NotificationRecipientAttrTypes = map[string]attr.Type{
	"id":                   types.StringType,
	"type":                 types.StringType,
	"target":               types.StringType,
	"notification_details": types.ListType{ElemType: types.ObjectType{AttrTypes: NotificationRecipientDetailsAttrTypes}},
}

var NotificationRecipientDetailsAttrTypes = map[string]attr.Type{
	"pagerduty_severity": types.StringType,
}
