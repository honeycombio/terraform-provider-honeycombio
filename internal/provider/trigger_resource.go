package provider

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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
								stringvalidator.OneOf(helper.TriggerThresholdOpStrings()...),
							},
						},
						"value": schema.Float64Attribute{
							Required:    true,
							Description: "The value to be used with the operator.",
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
			"recipient": schema.SetNestedBlock{
				Description:   "Zero or more recipients to notify when the Trigger fires.",
				PlanModifiers: []planmodifier.Set{modifiers.NotificationRecipients()},
				NestedObject: schema.NestedBlockObject{
					Validators: []validator.Object{
						objectvalidator.AtLeastOneOf(
							path.MatchRelative().AtName("id"),
							path.MatchRelative().AtName("type"),
						),
					},
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The ID of an existing recipient.",
							Validators: []validator.String{
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("type")),
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("target")),
							},
						},
						"type": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The type of the trigger recipient.",
							Validators: []validator.String{
								stringvalidator.OneOf(helper.RecipientTypeStrings(client.TriggerRecipientTypes())...),
							},
						},
						"target": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Target of the trigger recipient, this has another meaning depending on the type of recipient.",
							Validators: []validator.String{
								stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("type")),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"notification_details": schema.ListNestedBlock{
							Description: "Additional details to send along with the notification.",
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"pagerduty_severity": schema.StringAttribute{
										Optional:      true,
										Computed:      true,
										Description:   "The severity to set with the PagerDuty notification. If no severity is provided, 'critical' is assumed.",
										PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
										Validators: []validator.String{
											stringvalidator.All(
												stringvalidator.OneOf("info", "warning", "error", "critical"),
											),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *triggerResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	r.client = getClientFromResourceRequest(&req)
}

func (r *triggerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.TriggerResourceModel
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
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Honeycomb Trigger",
			"Could not create Trigger, unexpected error: "+err.Error(),
		)
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

	recipients := make([]models.NotificationRecipientModel, len(trigger.Recipients))
	for i, r := range trigger.Recipients {
		var rcpt models.NotificationRecipientModel

		// match the trigger's recipient to that in the plan
		idx := slices.IndexFunc(plan.Recipients, func(s models.NotificationRecipientModel) bool {
			if !s.ID.IsUnknown() {
				return s.ID.ValueString() == r.ID
			}
			return s.Type.ValueString() == string(r.Type) && (s.Target.IsNull() || s.Target.ValueString() == r.Target)
		})
		if idx < 0 {
			resp.Diagnostics.AddError(
				"Error Creating Honeycomb Trigger",
				"Could not find Recipient "+r.ID+" in plan",
			)
		}
		rcpt = plan.Recipients[idx]

		// TODO: can we move this to the planmodifier by adding a create state?
		if !rcpt.ID.IsUnknown() {
			// recipient provided by ID
			rcpt.ID = types.StringValue(r.ID)
			rcpt.Type = types.StringNull()
			rcpt.Target = types.StringNull()
		} else {
			// recipient provided by type+target
			rcpt.ID = types.StringNull()
			rcpt.Type = types.StringValue(string(r.Type))
			if rcpt.Type.ValueString() == string(client.RecipientTypePagerDuty) {
				// PagerDuty recipients don't have a target
				rcpt.Target = types.StringNull()
			} else {
				rcpt.Target = types.StringValue(r.Target)
			}
		}

		if r.Type == client.RecipientTypePagerDuty && r.Details != nil {
			rcpt.Details = make([]models.NotificationRecipientDetailsModel, 1)
			rcpt.Details[0].PDSeverity = types.StringValue(string(r.Details.PDSeverity))
		}
		recipients[i] = rcpt
	}
	state.Recipients = recipients

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *triggerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.TriggerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	trigger, err := r.client.Triggers.Get(ctx, state.Dataset.ValueString(), state.ID.ValueString())
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Honeycomb Trigger",
			"Could not read Trigger ID "+state.ID.ValueString()+": "+err.Error(),
		)
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
		for i, r := range trigger.Recipients {
			// match the Trigger's recipient to that in state
			idx := slices.IndexFunc(state.Recipients, func(s models.NotificationRecipientModel) bool {
				if !s.ID.IsNull() {
					return s.ID.ValueString() == r.ID
				}
				return s.Type.ValueString() == string(r.Type) && (s.Target.IsNull() || s.Target.ValueString() == r.Target)
			})
			if idx < 0 {
				resp.Diagnostics.AddError(
					"Error Reading Honeycomb Trigger",
					"Could not find Recipient "+r.ID+" in state",
				)
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
	var plan models.TriggerResourceModel
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
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Honeycomb Trigger",
			"Could not update Trigger, unexpected error: "+err.Error(),
		)
		return
	}

	trigger, err := r.client.Triggers.Get(ctx, plan.Dataset.ValueString(), plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Honeycomb Trigger",
			"Could not read Honeycomb Trigger ID "+plan.ID.ValueString()+": "+err.Error(),
		)
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

	recipients := make([]models.NotificationRecipientModel, len(trigger.Recipients))
	for i, r := range trigger.Recipients {
		// match the Trigger's recipient to that in the plan
		idx := slices.IndexFunc(plan.Recipients, func(s models.NotificationRecipientModel) bool {
			if !s.ID.IsNull() {
				return s.ID.ValueString() == r.ID
			}
			return s.Type.ValueString() == string(r.Type) && (s.Target.IsNull() || s.Target.ValueString() == r.Target)
		})
		if idx < 0 {
			resp.Diagnostics.AddError(
				"Error Updating Honeycomb Trigger",
				"Could not find Recipient "+r.ID+" in plan",
			)
		}
		recipients[i] = plan.Recipients[idx]
	}
	state.Recipients = recipients

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *triggerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.TriggerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Triggers.Delete(ctx, state.Dataset.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Honeycomb Trigger",
			"Could not delete Trigger, unexpected error: "+err.Error(),
		)
		return
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
		Op:    client.TriggerThresholdOp(t[0].Op.ValueString()),
		Value: t[0].Value.ValueFloat64(),
	}
}

func flattenTriggerThreshold(t *client.TriggerThreshold) []models.TriggerThresholdModel {
	return []models.TriggerThresholdModel{{
		Op:    types.StringValue(string(t.Op)),
		Value: types.Float64Value(t.Value),
	}}
}

func expandNotificationRecipients(n []models.NotificationRecipientModel) []client.NotificationRecipient {
	recipients := make([]client.NotificationRecipient, len(n))

	for i, r := range n {
		rcpt := client.NotificationRecipient{
			ID:     r.ID.ValueString(),
			Type:   client.RecipientType(r.Type.ValueString()),
			Target: r.Target.ValueString(),
		}
		if r.Details != nil {
			rcpt.Details = &client.NotificationRecipientDetails{
				PDSeverity: client.PagerDutySeverity(r.Details[0].PDSeverity.ValueString()),
			}
		}
		recipients[i] = rcpt
	}

	return recipients
}

func flattenNotificationRecipients(n []client.NotificationRecipient) []models.NotificationRecipientModel {
	recipients := make([]models.NotificationRecipientModel, len(n))

	for i, r := range n {
		rcpt := models.NotificationRecipientModel{
			ID:     types.StringValue(r.ID),
			Type:   types.StringValue(string(r.Type)),
			Target: types.StringValue(r.Target),
		}
		if r.Details != nil {
			rcpt.Details = make([]models.NotificationRecipientDetailsModel, 1)
			rcpt.Details[0].PDSeverity = types.StringValue(string(r.Details.PDSeverity))
		}
		recipients[i] = rcpt
	}

	return recipients
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
