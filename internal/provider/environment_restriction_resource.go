package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	hny "github.com/honeycombio/terraform-provider-honeycombio/client"
	v2client "github.com/honeycombio/terraform-provider-honeycombio/client/v2"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &environmentRestrictionResource{}
	_ resource.ResourceWithConfigure   = &environmentRestrictionResource{}
	_ resource.ResourceWithImportState = &environmentRestrictionResource{}
)

// environmentRestrictionResource restricts every dataset in an Environment
// while it exists. Deleting it removes the restriction.
type environmentRestrictionResource struct {
	client *v2client.Client
}

func NewEnvironmentRestrictionResource() resource.Resource {
	return &environmentRestrictionResource{}
}

func (*environmentRestrictionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_restriction"
}

func (r *environmentRestrictionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (*environmentRestrictionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Restricts every Dataset in a Honeycomb Environment. " +
			"Only Team owners and the members of a Group that has access to the Environment can read its Datasets.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the restriction. It is the ID of the Environment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Environment to restrict.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		},
	}
}

func (*environmentRestrictionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.Set(ctx, &models.EnvironmentRestrictionResourceModel{
		ID:            types.StringValue(req.ID),
		EnvironmentID: types.StringValue(req.ID),
	})...)
}

func (r *environmentRestrictionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.EnvironmentRestrictionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.EnvironmentID.ValueString()
	env, err := r.client.Environments.Get(ctx, id)
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.IsNotFound() {
		resp.Diagnostics.AddError(
			"Error Creating Honeycomb Environment Restriction",
			fmt.Sprintf("The API found no Environment %q.", id),
		)
		return
	}
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Creating Honeycomb Environment Restriction", err) {
		return
	}
	// A restriction made outside Terraform must be imported, so that
	// deleting this resource can't remove a restriction it didn't create.
	if env.Restricted != nil && *env.Restricted {
		resp.Diagnostics.AddError(
			"Honeycomb Environment Is Already Restricted",
			"To manage the existing restriction with Terraform, import it, "+
				"for example `terraform import honeycombio_environment_restriction.example "+id+"`.",
		)
		return
	}

	err = r.setRestricted(ctx, id, true)
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Creating Honeycomb Environment Restriction", err) {
		if errors.As(err, &detailedErr) && detailedErr.Status == http.StatusForbidden {
			// the API's 403 does not say why
			resp.Diagnostics.AddError(
				"The Management Key cannot restrict this Environment",
				"If a Group has access to a Dataset in the Environment, "+
					"the key also needs the `groups:write` scope to restrict it.",
			)
		}
		return
	}

	plan.ID = types.StringValue(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *environmentRestrictionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.EnvironmentRestrictionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env, err := r.client.Environments.Get(ctx, state.EnvironmentID.ValueString())
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.IsNotFound() {
		// the Environment is gone, and its restriction with it
		resp.State.RemoveResource(ctx)
		return
	}
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Reading Honeycomb Environment Restriction", err) {
		return
	}
	if env.Restricted == nil || !*env.Restricted {
		// the restriction was removed outside Terraform
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (*environmentRestrictionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.EnvironmentRestrictionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// every argument forces a replacement, so there is nothing to update
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *environmentRestrictionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.EnvironmentRestrictionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.setRestricted(ctx, state.EnvironmentID.ValueString(), false)
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.IsNotFound() {
		// the Environment is already gone
		return
	}
	helper.AddDiagnosticOnError(&resp.Diagnostics, "Deleting Honeycomb Environment Restriction", err)
}

func (r *environmentRestrictionResource) setRestricted(ctx context.Context, environmentID string, restricted bool) error {
	_, err := r.client.Environments.Update(ctx, &v2client.Environment{
		ID:         environmentID,
		Restricted: &restricted,
	})
	return err
}
