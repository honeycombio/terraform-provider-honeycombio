package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type BurnAlertResourceModel struct {
	ID                types.String                 `tfsdk:"id"`
	Dataset           types.String                 `tfsdk:"dataset"`
	SLOID             types.String                 `tfsdk:"slo_id"`
	ExhaustionMinutes types.Int64                  `tfsdk:"exhaustion_minutes"`
	Recipients        []NotificationRecipientModel `tfsdk:"recipient"`
}
