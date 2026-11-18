package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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
	_ resource.Resource                = &groupResource{}
	_ resource.ResourceWithConfigure   = &groupResource{}
	_ resource.ResourceWithImportState = &groupResource{}
)

type groupResource struct {
	client *v2client.Client
}

func NewGroupResource() resource.Resource {
	return &groupResource{}
}

func (*groupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *groupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (*groupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Group in a Honeycomb Team. A Group's members get the access the Group is granted.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Group. It must be unique within the Team.",
				Validators: []validator.String{
					stringvalidator.UTF8LengthBetween(1, 255),
				},
			},
			"description": schema.StringAttribute{
				Description: "The Group's description.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.UTF8LengthAtMost(1024),
				},
			},
		},
	}
}

func (r *groupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		resp.Diagnostics.AddError("Invalid Import ID", "The Group ID must be provided")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &models.GroupResourceModel{
		ID: types.StringValue(req.ID),
	})...)
}

func (r *groupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.GroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	g, err := r.client.Groups.Create(ctx, &v2client.Group{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueStringPointer(),
	})
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Creating Honeycomb Group", err) {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, flattenGroup(g))...)
}

func (r *groupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.GroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	g, err := r.client.Groups.Get(ctx, state.ID.ValueString())
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.IsNotFound() {
		// if not found consider it deleted -- so just remove it from state
		resp.State.RemoveResource(ctx)
		return
	}
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Reading Honeycomb Group", err) {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, flattenGroup(g))...)
}

func (r *groupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.GroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	g, err := r.client.Groups.Update(ctx, &v2client.Group{
		ID:          plan.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueStringPointer(),
	})
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Updating Honeycomb Group", err) {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, flattenGroup(g))...)
}

func (r *groupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.GroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Groups.Delete(ctx, state.ID.ValueString())
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.IsNotFound() {
		// already gone
		return
	}
	helper.AddDiagnosticOnError(&resp.Diagnostics, "Deleting Honeycomb Group", err)
}

func flattenGroup(g *v2client.Group) models.GroupResourceModel {
	var description string
	if g.Description != nil {
		description = *g.Description
	}
	return models.GroupResourceModel{
		ID:          types.StringValue(g.ID),
		Name:        types.StringValue(g.Name),
		Description: types.StringValue(description),
	}
}
