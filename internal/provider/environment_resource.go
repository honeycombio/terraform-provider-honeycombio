package provider

import (
	"context"
	"errors"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	hny "github.com/honeycombio/terraform-provider-honeycombio/client"
	v2client "github.com/honeycombio/terraform-provider-honeycombio/client/v2"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/modifiers"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &environmentResource{}
	_ resource.ResourceWithConfigure   = &environmentResource{}
	_ resource.ResourceWithModifyPlan  = &environmentResource{}
	_ resource.ResourceWithImportState = &environmentResource{}
)

type environmentResource struct {
	client *v2client.Client
}

func NewEnvironmentResource() resource.Resource {
	return &environmentResource{}
}

func (*environmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (r *environmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	w := getClientFromResourceRequest(&req)
	if w == nil {
		return
	}

	c, err := w.V2Client()
	if err != nil || c == nil {
		resp.Diagnostics.AddError("Failed to configure client", err.Error())
		return
	}
	r.client = c
}

func (*environmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Environment in a Honeycomb Team.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Required:    false,
				Optional:    false,
				Description: "The ID of the Environment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Environment.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"slug": schema.StringAttribute{
				Description: "The slug of the Environment.",
				Computed:    true,
				Optional:    false,
				Required:    false,
			},
			"description": schema.StringAttribute{
				Description: "The Environment's description.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
			"color": schema.StringAttribute{
				Description: "The color of the Environment. If one is not provided, a random color will be assigned.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(v2client.EnvironmentColorTypes()...),
				},
			},
			"delete_protected": schema.BoolAttribute{
				Description: "The current delete protection status of the Environment. Cannot be set to false on creation.",
				Default:     booldefault.StaticBool(true),
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					modifiers.EnforceDeletionProtection(),
				},
			},
		},
	}
}

func (r *environmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		resp.Diagnostics.AddError("Invalid Import ID", "The Environment ID must be provided")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &models.EnvironmentResourceModel{
		ID: types.StringValue(req.ID),
	})...)
}

func (r *environmentResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		// If the entire plan is null, the resource is planned for destruction -- let's add a warning
		resp.Diagnostics.AddWarning(
			"Resource Destruction Warning",
			"Appling this plan will delete the Environment and all of its contents. "+
				"This is an irreversible operation.",
		)
	}
}

func (r *environmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.EnvironmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env, err := r.client.Environments.Create(ctx, &v2client.Environment{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueStringPointer(),
		Color:       plan.Color.ValueStringPointer(),
	})
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Creating Environment", err) {
		return
	}

	var state models.EnvironmentResourceModel
	state.ID = types.StringValue(env.ID)
	state.Name = types.StringValue(env.Name)
	state.Slug = types.StringValue(env.Slug)
	state.Color = types.StringPointerValue(env.Color)
	state.Description = types.StringPointerValue(env.Description)
	state.DeleteProtected = types.BoolPointerValue(env.Settings.DeleteProtected)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *environmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.EnvironmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var detailedErr hny.DetailedError
	env, err := r.client.Environments.Get(ctx, state.ID.ValueString())
	if errors.As(err, &detailedErr) {
		if detailedErr.IsNotFound() {
			// if not found consider it deleted -- so just remove it from state
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.Append(helper.NewDetailedErrorDiagnostic(
				"Error Reading Honeycomb Environment",
				&detailedErr,
			))
		}
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Honeycomb Environment",
			"Unexpected error reading Environment ID "+state.ID.ValueString()+": "+err.Error(),
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(env.ID)
	state.Name = types.StringValue(env.Name)
	state.Slug = types.StringValue(env.Slug)
	state.Color = types.StringPointerValue(env.Color)
	state.Description = types.StringPointerValue(env.Description)
	state.DeleteProtected = types.BoolPointerValue(env.Settings.DeleteProtected)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *environmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state models.EnvironmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Environments.Update(ctx, &v2client.Environment{
		ID:          plan.ID.ValueString(),
		Description: plan.Description.ValueStringPointer(),
		Color:       plan.Color.ValueStringPointer(),
		Settings: &v2client.EnvironmentSettings{
			DeleteProtected: plan.DeleteProtected.ValueBoolPointer(),
		},
	})
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Updating Honeycomb Environment", err) {
		return
	}

	env, err := r.client.Environments.Get(ctx, plan.ID.ValueString())
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Updating Honeycomb Environment", err) {
		return
	}

	state.ID = types.StringValue(env.ID)
	state.Name = types.StringValue(env.Name)
	state.Slug = types.StringValue(env.Slug)
	state.Color = types.StringPointerValue(env.Color)
	state.Description = types.StringPointerValue(env.Description)
	state.DeleteProtected = types.BoolPointerValue(env.Settings.DeleteProtected)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *environmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.EnvironmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Environments.Delete(ctx, state.ID.ValueString())
	var detailedErr hny.DetailedError
	if err != nil {
		if errors.As(err, &detailedErr) {
			if detailedErr.Status == http.StatusConflict {
				resp.Diagnostics.AddError(
					"Unable to Delete Environment",
					"Delete Protection is enabled. "+
						"You must disable delete protection before it can be deleted.",
				)
			} else {
				resp.Diagnostics.Append(helper.NewDetailedErrorDiagnostic(
					"Error Deleting Honeycomb Environment",
					&detailedErr,
				))
			}
		} else {
			resp.Diagnostics.AddError(
				"Error Deleting Honeycomb Environment",
				"Could not delete Environment ID "+state.ID.ValueString()+": "+err.Error(),
			)
		}
	}
}
