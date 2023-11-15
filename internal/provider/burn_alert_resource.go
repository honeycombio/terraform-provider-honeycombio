package provider

import (
	"context"
	"errors"
	"strings"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/validation"
	"golang.org/x/exp/slices"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &burnAlertResource{}
	_ resource.ResourceWithConfigure      = &burnAlertResource{}
	_ resource.ResourceWithImportState    = &burnAlertResource{}
	_ resource.ResourceWithValidateConfig = &burnAlertResource{}
)

type burnAlertResource struct {
	client *client.Client
}

func NewBurnAlertResource() resource.Resource {
	return &burnAlertResource{}
}

func (*burnAlertResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_burn_alert"
}

func (r *burnAlertResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = getClientFromResourceRequest(&req)
}

func (*burnAlertResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Burn Alerts are used to notify you when your error budget will be exhausted within a given time period.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this Burn Alert.",
				Computed:    true,
				Required:    false,
				Optional:    false,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"alert_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The alert type of this Burn Alert.",
				Default:     stringdefault.StaticString(string(client.BurnAlertAlertTypeExhaustionTime)),
				Validators: []validator.String{
					stringvalidator.OneOf(helper.AsStringSlice(client.BurnAlertAlertTypes())...),
				},
			},
			"budget_rate_decrease_percent": schema.Float64Attribute{
				Optional:    true,
				Description: "The percent the budget has decreased over the budget rate window.",
				Validators: []validator.Float64{
					float64validator.AtLeast(0.0001),
					float64validator.AtMost(100),
					validation.Float64PrecisionAtMost(4),
				},
			},
			"budget_rate_window_minutes": schema.Int64Attribute{
				Optional:    true,
				Description: "The time period, in minutes, over which a budget rate will be calculated.",
				Validators: []validator.Int64{
					int64validator.AtLeast(60),
				},
			},
			"dataset": schema.StringAttribute{
				Required:    true,
				Description: "The dataset this Burn Alert is associated with.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"exhaustion_minutes": schema.Int64Attribute{
				Optional:    true,
				Description: "The amount of time, in minutes, remaining before the SLO's error budget will be exhausted and the alert will fire.",
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"slo_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the SLO that this Burn Alert is associated with.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"recipient": notificationRecipientSchema(client.RecipientTypes()),
		},
	}
}

func validateAttributesWhenAlertTypeIsExhaustionTime(data models.BurnAlertResourceModel, resp *resource.ValidateConfigResponse) {
	// Check that the alert_type is exhaustion_time or that it is not configured(which means we default to exhaustion_time)
	if data.AlertType.IsNull() || data.AlertType.IsUnknown() || data.AlertType.ValueString() == string(client.BurnAlertAlertTypeExhaustionTime) {
		// When the alert_type is exhaustion_time, exhaustion_minutes is required
		if data.ExhaustionMinutes.IsNull() || data.ExhaustionMinutes.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("exhaustion_minutes"),
				"Missing required argument",
				"The argument \"exhaustion_minutes\" is required, but no definition was found.",
			)
		}

		// When the alert_type is exhaustion_time, budget_rate_decrease_percent must not be configured
		if !(data.BudgetRateDecreasePercent.IsNull() || data.BudgetRateDecreasePercent.IsUnknown()) {
			resp.Diagnostics.AddAttributeError(
				path.Root("budget_rate_decrease_percent"),
				"Conflicting configuration arguments",
				"\"budget_rate_decrease_percent\": must not be configured when \"alert_type\" is \"exhaustion_time\"",
			)
		}

		// When the alert_type is exhaustion_time, budget_rate_window_minutes must not be configured
		if !(data.BudgetRateWindowMinutes.IsNull() || data.BudgetRateWindowMinutes.IsUnknown()) {
			resp.Diagnostics.AddAttributeError(
				path.Root("budget_rate_window_minutes"),
				"Conflicting configuration arguments",
				"\"budget_rate_window_minutes\": must not be configured when \"alert_type\" is \"exhaustion_time\"",
			)
		}
	}
}

func validateAttributesWhenAlertTypeIsBudgetRate(data models.BurnAlertResourceModel, resp *resource.ValidateConfigResponse) {
	// Check if the alert_type is budget_rate
	if data.AlertType.ValueString() == string(client.BurnAlertAlertTypeBudgetRate) {
		// When the alert_type is budget_rate, budget_rate_decrease_percent is required
		if data.BudgetRateDecreasePercent.IsNull() || data.BudgetRateDecreasePercent.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("budget_rate_decrease_percent"),
				"Missing required argument",
				"The argument \"budget_rate_decrease_percent\" is required, but no definition was found.",
			)
		}

		// When the alert_type is budget_rate, budget_rate_window_minutes is required
		if data.BudgetRateWindowMinutes.IsNull() || data.BudgetRateWindowMinutes.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("budget_rate_window_minutes"),
				"Missing required argument",
				"The argument \"budget_rate_window_minutes\" is required, but no definition was found.",
			)
		}

		// When the alert_type is budget_rate, exhaustion_minutes must not be configured
		if !(data.ExhaustionMinutes.IsNull() || data.ExhaustionMinutes.IsUnknown()) {
			resp.Diagnostics.AddAttributeError(
				path.Root("exhaustion_minutes"),
				"Conflicting configuration arguments",
				"\"exhaustion_minutes\": must not be configured when \"alert_type\" is \"budget_rate\"",
			)
		}
	}
}

func (r *burnAlertResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data models.BurnAlertResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// When alert_type is exhaustion_time, check that exhaustion_minutes
	// is configured and budget rate attributes are not configured
	validateAttributesWhenAlertTypeIsExhaustionTime(data, resp)

	// When alert_type is budget_rate, check that budget rate
	// attributes are configured and exhaustion_minutes is not configured
	validateAttributesWhenAlertTypeIsBudgetRate(data, resp)

	return
}

func (r *burnAlertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// import ID is of the format <dataset>/<BurnAlert ID>
	// note that the dataset name can also contain '/'
	idSegments := strings.Split(req.ID, "/")
	if len(idSegments) < 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"The supplied ID must be written as <dataset>/<BurnAlert ID>.",
		)
		return
	}

	id := idSegments[len(idSegments)-1]
	dataset := strings.Join(idSegments[0:len(idSegments)-1], "/")

	resp.Diagnostics.Append(resp.State.Set(ctx, &models.BurnAlertResourceModel{
		ID:      types.StringValue(id),
		Dataset: types.StringValue(dataset),
	})...)
}

func (r *burnAlertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config models.BurnAlertResourceModel
	// Read in the config and plan data
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get attributes from config and construct the create request
	createRequest := &client.BurnAlert{
		AlertType:  plan.AlertType.ValueString(),
		Recipients: expandNotificationRecipients(plan.Recipients),
		SLO:        client.SLORef{ID: plan.SLOID.ValueString()},
	}

	// Process any attributes that could be nil and add them to the create request
	exhaustionMinutes := int(plan.ExhaustionMinutes.ValueInt64())
	// Must convert from float to PPM because the API only accepts PPM
	budgetRateDecreasePercentAsPPM := helper.FloatToPPM(plan.BudgetRateDecreasePercent.ValueFloat64())
	budgetRateWindowMinutes := int(plan.BudgetRateWindowMinutes.ValueInt64())
	if plan.AlertType.ValueString() == string(client.BurnAlertAlertTypeExhaustionTime) {
		createRequest.ExhaustionMinutes = &exhaustionMinutes
	}
	if plan.AlertType.ValueString() == string(client.BurnAlertAlertTypeBudgetRate) {
		createRequest.BudgetRateDecreaseThresholdPerMillion = &budgetRateDecreasePercentAsPPM
		createRequest.BudgetRateWindowMinutes = &budgetRateWindowMinutes
	}

	// Create the new burn alert
	burnAlert, err := r.client.BurnAlerts.Create(ctx, plan.Dataset.ValueString(), createRequest)
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Creating Honeycomb Burn Alert", err) {
		return
	}

	// Get attributes from the new burn alert and construct the state values
	var state models.BurnAlertResourceModel
	state.ID = types.StringValue(burnAlert.ID)
	state.AlertType = types.StringValue(burnAlert.AlertType)
	state.Dataset = plan.Dataset
	// we created them as authored so to avoid matching type-target or ID we can just use the same value
	state.Recipients = config.Recipients
	state.SLOID = types.StringValue(burnAlert.SLO.ID)

	// Process any attributes that could be nil and add them to the state values
	if burnAlert.ExhaustionMinutes != nil {
		state.ExhaustionMinutes = types.Int64Value(int64(*burnAlert.ExhaustionMinutes))
	}
	if burnAlert.BudgetRateDecreaseThresholdPerMillion != nil {
		// Must convert from PPM back to float to match what the user has in their config
		budgetRateDecreaseThresholdPerMillionAsPercent := helper.PPMToFloat(*burnAlert.BudgetRateDecreaseThresholdPerMillion)
		state.BudgetRateDecreasePercent = types.Float64Value(budgetRateDecreaseThresholdPerMillionAsPercent)
	}
	if burnAlert.BudgetRateWindowMinutes != nil {
		state.BudgetRateWindowMinutes = types.Int64Value(int64(*burnAlert.BudgetRateWindowMinutes))
	}

	// Set the new burn alert's attributes in state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *burnAlertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.BurnAlertResourceModel
	// Read in the state data
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the burn alert, using the values from state
	var detailedErr *client.DetailedError
	burnAlert, err := r.client.BurnAlerts.Get(ctx, state.Dataset.ValueString(), state.ID.ValueString())
	if errors.As(err, &detailedErr) {
		if detailedErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.Append(helper.NewDetailedErrorDiagnostic(
				"Error Reading Honeycomb Burn Alert",
				detailedErr,
			))
		}
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Honeycomb Burn Alert",
			"Could not read Burn Alert ID "+state.ID.ValueString()+": "+err.Error(),
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Get attributes from the burn alert and construct the state values
	state.ID = types.StringValue(burnAlert.ID)
	state.AlertType = types.StringValue(burnAlert.AlertType)
	state.SLOID = types.StringValue(burnAlert.SLO.ID)

	// Process any attributes that could be nil and add them to the state values
	if burnAlert.ExhaustionMinutes != nil {
		state.ExhaustionMinutes = types.Int64Value(int64(*burnAlert.ExhaustionMinutes))
	}
	if burnAlert.BudgetRateDecreaseThresholdPerMillion != nil {
		// Must convert from PPM back to float to match what the user has in their config
		budgetRateDecreaseThresholdPerMillionAsPercent := helper.PPMToFloat(*burnAlert.BudgetRateDecreaseThresholdPerMillion)
		state.BudgetRateDecreasePercent = types.Float64Value(budgetRateDecreaseThresholdPerMillionAsPercent)
	}
	if burnAlert.BudgetRateWindowMinutes != nil {
		state.BudgetRateWindowMinutes = types.Int64Value(int64(*burnAlert.BudgetRateWindowMinutes))
	}

	recipients := make([]models.NotificationRecipientModel, len(burnAlert.Recipients))
	if state.Recipients != nil {
		// match the recipients to those in the state sorting out type+target vs ID
		for i, r := range burnAlert.Recipients {
			idx := slices.IndexFunc(state.Recipients, func(s models.NotificationRecipientModel) bool {
				if !s.ID.IsNull() {
					return s.ID.ValueString() == r.ID
				}
				return s.Type.ValueString() == string(r.Type) && s.Target.ValueString() == r.Target
			})
			if idx < 0 {
				// this should never happen?! But if it does, we'll just skip it and hope to get a reproducable case
				resp.Diagnostics.AddError(
					"Error Reading Honeycomb Burn Alert",
					"Could not find Recipient "+r.ID+" in state",
				)
				continue
			}
			recipients[i] = state.Recipients[idx]
		}
	} else {
		recipients = flattenNotificationRecipients(burnAlert.Recipients)
	}
	state.Recipients = recipients

	// Set the burn alert's attributes in state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *burnAlertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, config models.BurnAlertResourceModel
	// Read in the config and plan data
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get attributes from config and construct the update request
	updateRequest := &client.BurnAlert{
		ID:         plan.ID.ValueString(),
		AlertType:  plan.AlertType.ValueString(),
		Recipients: expandNotificationRecipients(plan.Recipients),
		SLO:        client.SLORef{ID: plan.SLOID.ValueString()},
	}

	// Process any attributes that could be nil and add them to the update request
	exhaustionMinutes := int(plan.ExhaustionMinutes.ValueInt64())
	// Must convert from float to PPM because the API only accepts PPM
	budgetRateDecreasePercentAsPPM := helper.FloatToPPM(plan.BudgetRateDecreasePercent.ValueFloat64())
	budgetRateWindowMinutes := int(plan.BudgetRateWindowMinutes.ValueInt64())
	if plan.AlertType.ValueString() == string(client.BurnAlertAlertTypeExhaustionTime) {
		updateRequest.ExhaustionMinutes = &exhaustionMinutes
	}
	if plan.AlertType.ValueString() == string(client.BurnAlertAlertTypeBudgetRate) {
		updateRequest.BudgetRateDecreaseThresholdPerMillion = &budgetRateDecreasePercentAsPPM
		updateRequest.BudgetRateWindowMinutes = &budgetRateWindowMinutes
	}

	// Update the burn alert
	_, err := r.client.BurnAlerts.Update(ctx, plan.Dataset.ValueString(), updateRequest)
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Updating Honeycomb Burn Alert", err) {
		return
	}

	// Read the updated burn alert
	burnAlert, err := r.client.BurnAlerts.Get(ctx, plan.Dataset.ValueString(), plan.ID.ValueString())
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Updating Honeycomb Burn Alert", err) {
		return
	}

	// Get attributes from the updated burn alert and construct the state values
	var state models.BurnAlertResourceModel
	state.ID = types.StringValue(burnAlert.ID)
	state.AlertType = types.StringValue(burnAlert.AlertType)
	state.Dataset = plan.Dataset
	// we created them as authored so to avoid matching type-target or ID we can just use the same value
	state.Recipients = config.Recipients
	state.SLOID = types.StringValue(burnAlert.SLO.ID)

	// Process any attributes that could be nil and add them to the state values
	if burnAlert.ExhaustionMinutes != nil {
		state.ExhaustionMinutes = types.Int64Value(int64(*burnAlert.ExhaustionMinutes))
	}
	if burnAlert.BudgetRateDecreaseThresholdPerMillion != nil {
		// Must convert from PPM back to float to match what the user has in their config
		budgetRateDecreaseThresholdPerMillionAsPercent := helper.PPMToFloat(*burnAlert.BudgetRateDecreaseThresholdPerMillion)
		state.BudgetRateDecreasePercent = types.Float64Value(budgetRateDecreaseThresholdPerMillionAsPercent)
	}
	if burnAlert.BudgetRateWindowMinutes != nil {
		state.BudgetRateWindowMinutes = types.Int64Value(int64(*burnAlert.BudgetRateWindowMinutes))
	}

	// Set the updated burn alert's attributes in state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *burnAlertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.BurnAlertResourceModel
	// Read in the state data
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the burn alert, using the values from state
	err := r.client.BurnAlerts.Delete(ctx, state.Dataset.ValueString(), state.ID.ValueString())
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Deleting Honeycomb Burn Alert", err) {
		return
	}
}
