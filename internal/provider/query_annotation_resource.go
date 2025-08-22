package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/modifiers"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &queryAnnotationResource{}
	_ resource.ResourceWithConfigure = &queryAnnotationResource{}
)

type queryAnnotationResource struct {
	client *client.Client
}

func NewQueryAnnotationResource() resource.Resource {
	return &queryAnnotationResource{}
}

func (*queryAnnotationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_query_annotation"
}

func (r *queryAnnotationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	w := getClientFromResourceRequest(&req)
	if w == nil {
		return
	}

	c, err := w.V1Client()
	if err != nil || c == nil {
		resp.Diagnostics.AddError("Failed to configure client", err.Error())
		return
	}
	r.client = c
}

func (*queryAnnotationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Query annotations allow you to add notes and descriptions to specific queries for better documentation and context.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this Query Annotation.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dataset": schema.StringAttribute{
				Description: "The dataset this query annotation is added to. If not set, an Environment-wide query annotation will be created.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					modifiers.DatasetDeprecation(true),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"query_id": schema.StringAttribute{
				Description: "The ID of the query that the annotation will be created on. Note that a query can have more than one annotation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name to display as the query annotation.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 320),
				},
			},
			"description": schema.StringAttribute{
				Description: "The description to display as the query annotation.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1023),
				},
			},
		},
	}
}

func (r *queryAnnotationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.QueryAnnotationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataset := helper.GetDatasetOrAll(plan.Dataset)
	queryAnnotation := &client.QueryAnnotation{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		QueryID:     plan.QueryID.ValueString(),
	}

	createdAnnotation, err := r.client.QueryAnnotations.Create(ctx, dataset.ValueString(), queryAnnotation)
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Creating Honeycomb Query Annotation", err) {
		return
	}

	var state models.QueryAnnotationResourceModel
	state.ID = types.StringValue(createdAnnotation.ID)
	state.Dataset = plan.Dataset
	state.QueryID = plan.QueryID
	state.Name = types.StringValue(createdAnnotation.Name)
	state.Description = types.StringValue(createdAnnotation.Description)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *queryAnnotationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.QueryAnnotationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataset := helper.GetDatasetOrAll(state.Dataset)

	var detailedErr client.DetailedError
	queryAnnotation, err := r.client.QueryAnnotations.Get(ctx, dataset.ValueString(), state.ID.ValueString())
	if errors.As(err, &detailedErr) {
		if detailedErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(helper.NewDetailedErrorDiagnostic("Reading Honeycomb Query Annotation", &detailedErr))
		return
	}
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Reading Honeycomb Query Annotation", err) {
		return
	}

	state.Name = types.StringValue(queryAnnotation.Name)
	state.Description = types.StringValue(queryAnnotation.Description)
	state.QueryID = types.StringValue(queryAnnotation.QueryID)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *queryAnnotationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.QueryAnnotationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataset := helper.GetDatasetOrAll(plan.Dataset)
	queryAnnotation := &client.QueryAnnotation{
		ID:          plan.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		QueryID:     plan.QueryID.ValueString(),
	}

	updatedAnnotation, err := r.client.QueryAnnotations.Update(ctx, dataset.ValueString(), queryAnnotation)
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Updating Honeycomb Query Annotation", err) {
		return
	}

	// Update the state with the returned values
	var state models.QueryAnnotationResourceModel
	state.ID = plan.ID
	state.Dataset = plan.Dataset
	state.QueryID = plan.QueryID
	state.Name = types.StringValue(updatedAnnotation.Name)
	state.Description = types.StringValue(updatedAnnotation.Description)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *queryAnnotationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.QueryAnnotationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataset := state.Dataset.ValueString()
	if dataset == "" {
		dataset = client.EnvironmentWideSlug
	}

	err := r.client.QueryAnnotations.Delete(ctx, dataset, state.ID.ValueString())
	helper.AddDiagnosticOnError(&resp.Diagnostics, "Deleting Honeycomb Query Annotation", err)
}
