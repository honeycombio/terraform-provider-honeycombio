package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type NotificationRecipientModel struct {
	ID      types.String                        `tfsdk:"id"`
	Type    types.String                        `tfsdk:"type"`
	Target  types.String                        `tfsdk:"target"`
	Details []NotificationRecipientDetailsModel `tfsdk:"notification_details"`
}

type NotificationRecipientDetailsModel struct {
	PDSeverity types.String `tfsdk:"pagerduty_severity"`
}
