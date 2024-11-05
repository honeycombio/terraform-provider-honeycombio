package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type BurnAlertResourceModel struct {
	ID                        types.String  `tfsdk:"id"`
	AlertType                 types.String  `tfsdk:"alert_type"`
	BudgetRateWindowMinutes   types.Int64   `tfsdk:"budget_rate_window_minutes"`
	BudgetRateDecreasePercent types.Float64 `tfsdk:"budget_rate_decrease_percent"`
	Description               types.String  `tfsdk:"description"`
	Dataset                   types.String  `tfsdk:"dataset"`
	SLOID                     types.String  `tfsdk:"slo_id"`
	ExhaustionMinutes         types.Int64   `tfsdk:"exhaustion_minutes"`
	Recipients                types.Set     `tfsdk:"recipient"` // NotificationRecipientModel
}
