package provider

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/modifiers"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/validation"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &triggerResource{}
	_ resource.ResourceWithConfigure   = &triggerResource{}
	_ resource.ResourceWithImportState = &triggerResource{}
)

func NewTriggerResource() resource.Resource {
	return &triggerResource{}
}

type triggerResource struct {
	client *client.Client
}

// matches HH:mm timestamps with optional leading 0
//
//	e.g. 9:00, 09:01
var hhMMRegex = regexp.MustCompile(`^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$`)

func (r *triggerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trigger"
}

func (r *triggerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Trigger.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this Trigger.",
				Computed:    true,
				Required:    false,
				Optional:    false,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Trigger.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 120),
				},
			},
			"dataset": schema.StringAttribute{
				Required:    true,
				Description: "The dataset this Trigger is associated with.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "A description of the Trigger.",
				Default:     stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1023),
				},
			},
			"disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The state of the Trigger. If true, the Trigger will not be run.",
				Default:     booldefault.StaticBool(false),
			},
			"query_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Query that the Trigger will execute.",
			},
			"alert_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Control when the Trigger will send a notification.",
				Default:     stringdefault.StaticString(string(client.TriggerAlertTypeOnChange)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(client.TriggerAlertTypeOnChange),
						string(client.TriggerAlertTypeOnTrue),
					),
				},
			},
			"frequency": schema.Int64Attribute{
				Optional: true,
				Description: "The interval (in seconds) in which to check the results of the query's calculation against the threshold. " +
					"This value must be divisible by 60, between 60 and 86400 (between 1 minute and 1 day), and not be more than 4 times the query's duration.",
				Computed: true,
				Default:  int64default.StaticInt64(900),
				Validators: []validator.Int64{
					int64validator.All(
						validation.Int64DivisibleBy(60),
						int64validator.Between(60, 86400),
					),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"threshold": schema.ListNestedBlock{
				Description: "A block describing the threshold for the Trigger to fire.",
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"op": schema.StringAttribute{
							Required:    true,
							Description: "The operator to apply.",
							Validators: []validator.String{
								stringvalidator.OneOf(helper.AsStringSlice(client.TriggerThresholdOps())...),
							},
						},
						"value": schema.Float64Attribute{
							Required:    true,
							Description: "The value to be used with the operator.",
						},
						"exceeded_limit": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Description: "The number of times the threshold is met before an alert is sent. Defaults to 1.",
							Validators:  []validator.Int64{int64validator.Between(1, 5)},
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
								// Nested blocks currently do not support default values, so we need to set it via a planmodifier
								modifiers.DefaultTriggerThresholdExceededLimit(),
							},
						},
					},
				},
			},
			"evaluation_schedule": schema.ListNestedBlock{
				Description: "The schedule that determines when the trigger is run. When the time is within the scheduled window, " +
					" the trigger will be run at the specified frequency. Outside of the window, the trigger will not be run." +
					"If no schedule is specified, the trigger will be run at the specified frequency at all times.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"start_time": schema.StringAttribute{
							Description: "UTC time to start evaluating the trigger in HH:mm format",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(hhMMRegex, "must be in HH:mm format"),
							},
						},
						"end_time": schema.StringAttribute{
							Description: "UTC time to stop evaluating the trigger in HH:mm format",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(hhMMRegex, "must be in HH:mm format"),
							},
						},
						"days_of_week": schema.ListAttribute{
							ElementType: types.StringType,
							Description: "The days of the week to evaluate the trigger on",
							Required:    true,
							Validators: []validator.List{
								listvalidator.ValueStringsAre(stringvalidator.OneOf("monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday")),
								listvalidator.UniqueValues(),
							},
						},
					},
				},
			},
			"recipient": notificationRecipientSchema(client.TriggerRecipientTypes()),
		},
	}
}

func (r *triggerResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	r.client = getClientFromResourceRequest(&req)
}

func (r *triggerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config models.TriggerResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	newTrigger := &client.Trigger{
		Name:               plan.Name.ValueString(),
		Description:        plan.Description.ValueString(),
		Disabled:           plan.Disabled.ValueBool(),
		QueryID:            plan.QueryID.ValueString(),
		AlertType:          client.TriggerAlertType(plan.AlertType.ValueString()),
		Threshold:          expandTriggerThreshold(plan.Threshold),
		Frequency:          int(plan.Frequency.ValueInt64()),
		Recipients:         expandNotificationRecipients(plan.Recipients),
		EvaluationSchedule: expandTriggerEvaluationSchedule(plan.EvaluationSchedule),
	}
	if plan.EvaluationSchedule != nil {
		newTrigger.EvaluationScheduleType = client.TriggerEvaluationScheduleWindow
	}

	trigger, err := r.client.Triggers.Create(ctx, plan.Dataset.ValueString(), newTrigger)
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Creating Honeycomb Trigger", err) {
		return
	}

	var state models.TriggerResourceModel
	state.Dataset = plan.Dataset
	state.ID = types.StringValue(trigger.ID)
	state.Name = types.StringValue(trigger.Name)
	state.Description = types.StringValue(trigger.Description)
	state.Disabled = types.BoolValue(trigger.Disabled)
	state.QueryID = types.StringValue(trigger.QueryID)
	state.AlertType = types.StringValue(string(trigger.AlertType))
	state.Threshold = flattenTriggerThreshold(trigger.Threshold)
	state.Frequency = types.Int64Value(int64(trigger.Frequency))
	state.EvaluationSchedule = flattenTriggerEvaluationSchedule(trigger)
	// we created them as authored so to avoid matching type-target or ID we can just use the same value
	state.Recipients = config.Recipients

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *triggerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.TriggerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var detailedErr client.DetailedError
	trigger, err := r.client.Triggers.Get(ctx, state.Dataset.ValueString(), state.ID.ValueString())
	if errors.As(err, &detailedErr) {
		if detailedErr.IsNotFound() {
			// if not found consider it deleted -- so just remove it from state
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.Append(helper.NewDetailedErrorDiagnostic(
				"Error Reading Honeycomb Trigger",
				&detailedErr,
			))
		}
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Honeycomb Trigger",
			"Unexpected error reading Trigger ID "+state.ID.ValueString()+": "+err.Error(),
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(trigger.ID)
	state.Name = types.StringValue(trigger.Name)
	state.Description = types.StringValue(trigger.Description)
	state.Disabled = types.BoolValue(trigger.Disabled)
	state.QueryID = types.StringValue(trigger.QueryID)
	state.AlertType = types.StringValue(string(trigger.AlertType))
	state.Threshold = flattenTriggerThreshold(trigger.Threshold)
	state.Frequency = types.Int64Value(int64(trigger.Frequency))
	state.EvaluationSchedule = flattenTriggerEvaluationSchedule(trigger)

	recipients := make([]models.NotificationRecipientModel, len(trigger.Recipients))
	if state.Recipients != nil {
		// match the Trigger's recipients to those in the state sorting out type+target vs ID
		for i, r := range trigger.Recipients {
			idx := slices.IndexFunc(state.Recipients, func(s models.NotificationRecipientModel) bool {
				if !s.ID.IsNull() {
					return s.ID.ValueString() == r.ID
				}
				return s.Type.ValueString() == string(r.Type) && s.Target.ValueString() == r.Target
			})
			if idx < 0 {
				// this should never happen?! But if it does, we'll just skip it and hope to get a reproducable case
				resp.Diagnostics.AddError(
					"Error Reading Honeycomb Trigger",
					"Could not find Recipient "+r.ID+" in state",
				)
				continue
			}
			recipients[i] = state.Recipients[idx]
		}
	} else {
		recipients = flattenNotificationRecipients(trigger.Recipients)
	}
	state.Recipients = recipients

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *triggerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, config models.TriggerResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updatedTrigger := &client.Trigger{
		ID:                 plan.ID.ValueString(),
		Name:               plan.Name.ValueString(),
		Description:        plan.Description.ValueString(),
		Disabled:           plan.Disabled.ValueBool(),
		QueryID:            plan.QueryID.ValueString(),
		AlertType:          client.TriggerAlertType(plan.AlertType.ValueString()),
		Frequency:          int(plan.Frequency.ValueInt64()),
		Threshold:          expandTriggerThreshold(plan.Threshold),
		Recipients:         expandNotificationRecipients(plan.Recipients),
		EvaluationSchedule: expandTriggerEvaluationSchedule(plan.EvaluationSchedule),
	}
	if updatedTrigger.EvaluationSchedule != nil {
		updatedTrigger.EvaluationScheduleType = client.TriggerEvaluationScheduleWindow
	} else {
		updatedTrigger.EvaluationScheduleType = client.TriggerEvaluationScheduleFrequency
	}

	_, err := r.client.Triggers.Update(ctx, plan.Dataset.ValueString(), updatedTrigger)
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Updating Honeycomb Trigger", err) {
		return
	}

	trigger, err := r.client.Triggers.Get(ctx, plan.Dataset.ValueString(), plan.ID.ValueString())
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Updating Honeycomb Trigger", err) {
		return
	}

	var state models.TriggerResourceModel
	state.Dataset = plan.Dataset
	state.ID = types.StringValue(trigger.ID)
	state.Name = types.StringValue(trigger.Name)
	state.Description = types.StringValue(trigger.Description)
	state.Disabled = types.BoolValue(trigger.Disabled)
	state.QueryID = types.StringValue(trigger.QueryID)
	state.AlertType = types.StringValue(string(trigger.AlertType))
	state.Frequency = types.Int64Value(int64(trigger.Frequency))
	state.Threshold = flattenTriggerThreshold(trigger.Threshold)
	state.EvaluationSchedule = flattenTriggerEvaluationSchedule(trigger)
	// we created them as authored so to avoid matching type-target or ID we can just use the same value
	state.Recipients = config.Recipients

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *triggerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.TriggerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var detailedErr client.DetailedError
	err := r.client.Triggers.Delete(ctx, state.Dataset.ValueString(), state.ID.ValueString())
	if errors.As(err, &detailedErr) {
		// if not found consider it deleted -- so don't error
		if !detailedErr.IsNotFound() {
			resp.Diagnostics.Append(helper.NewDetailedErrorDiagnostic(
				"Error Deleting Honeycomb Trigger",
				&detailedErr,
			))
		}
	} else {
		resp.Diagnostics.AddError(
			"Error Deleting Honeycomb Trigger",
			"Could not delete Trigger ID "+state.ID.ValueString()+": "+err.Error(),
		)
	}
}

func (r *triggerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// import ID is of the format <dataset>/<trigger ID>
	// note that the dataset name can also contain '/'
	idSegments := strings.Split(req.ID, "/")
	if len(idSegments) < 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"The supplied ID must be wrtten as <dataset>/<trigger ID>.",
		)
		return
	}

	id := idSegments[len(idSegments)-1]
	dataset := strings.Join(idSegments[0:len(idSegments)-1], "/")

	resp.Diagnostics.Append(resp.State.Set(ctx, &models.TriggerResourceModel{
		ID:      types.StringValue(id),
		Dataset: types.StringValue(dataset),
	})...)
}

func expandTriggerThreshold(t []models.TriggerThresholdModel) *client.TriggerThreshold {
	if len(t) != 1 {
		return nil
	}

	return &client.TriggerThreshold{
		Op:            client.TriggerThresholdOp(t[0].Op.ValueString()),
		Value:         t[0].Value.ValueFloat64(),
		ExceededLimit: int(t[0].ExceededLimit.ValueInt64()),
	}
}

func flattenTriggerThreshold(t *client.TriggerThreshold) []models.TriggerThresholdModel {
	return []models.TriggerThresholdModel{{
		Op:            types.StringValue(string(t.Op)),
		Value:         types.Float64Value(t.Value),
		ExceededLimit: types.Int64Value(int64(t.ExceededLimit)),
	}}
}

func expandTriggerEvaluationSchedule(s []models.TriggerEvaluationScheduleModel) *client.TriggerEvaluationSchedule {
	if s != nil {
		days := make([]string, len(s[0].DaysOfWeek))
		for i, d := range s[0].DaysOfWeek {
			days[i] = d.ValueString()
		}

		return &client.TriggerEvaluationSchedule{
			Window: client.TriggerEvaluationWindow{
				StartTime:  s[0].StartTime.ValueString(),
				EndTime:    s[0].EndTime.ValueString(),
				DaysOfWeek: days,
			},
		}
	}

	return nil
}

func flattenTriggerEvaluationSchedule(t *client.Trigger) []models.TriggerEvaluationScheduleModel {
	if t.EvaluationScheduleType == client.TriggerEvaluationScheduleWindow {
		days := make([]basetypes.StringValue, len(t.EvaluationSchedule.Window.DaysOfWeek))
		for i, d := range t.EvaluationSchedule.Window.DaysOfWeek {
			days[i] = types.StringValue(d)
		}

		return []models.TriggerEvaluationScheduleModel{
			{
				StartTime:  types.StringValue(t.EvaluationSchedule.Window.StartTime),
				EndTime:    types.StringValue(t.EvaluationSchedule.Window.EndTime),
				DaysOfWeek: days,
			},
		}
	}

	return nil
}
