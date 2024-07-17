package provider

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/modifiers"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/validation"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &queryResource{}
	_ resource.ResourceWithConfigure   = &queryResource{}
	_ resource.ResourceWithImportState = &queryResource{}
)

type queryResource struct {
	client *client.Client
}

func NewQueryResource() resource.Resource {
	return &queryResource{}
}

func (*queryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_query"
}

func (r *queryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (*queryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Queries can be used by Triggers and Boards, or be executed via the Query Data API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this Query.",
				Computed:    true,
				Required:    false,
				Optional:    false,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dataset": schema.StringAttribute{
				Description: "The dataset this Query is associated with. Use `__all__` for envionment-wide queries.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"query_json": schema.StringAttribute{
				Description: "A JSON object describing the query according to the Query Specification." +
					" While the JSON can be constructed manually, it is easiest to use the `honeycombio_query_specification` data source.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					modifiers.EquivalentQuerySpec(),
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validation.ValidQuerySpec(),
				},
			},
		},
	}
}

func (r *queryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// import ID is of the format <dataset>/<query ID>
	dataset, id, found := strings.Cut(req.ID, "/")
	if !found {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"The supplied ID must be wrtten as <dataset>/<query ID>.",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &models.QueryResourceModel{
		ID:      types.StringValue(id),
		Dataset: types.StringValue(dataset),
	})...)
}

func (r *queryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.QueryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var querySpec client.QuerySpec
	if err := json.Unmarshal([]byte(plan.QueryJson.ValueString()), &querySpec); err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("query_json"), "Failed to unmarshal JSON", err.Error())
		return
	}
	query, err := r.client.Queries.Create(ctx, plan.Dataset.ValueString(), &querySpec)
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Creating Honeycomb Query", err) {
		return
	}

	var state models.QueryResourceModel
	state.ID = types.StringValue(*query.ID)
	state.Dataset = plan.Dataset
	// store the plan's query JSON in state so it matches the config and rely on the plan modifier
	// to handle the rest when we read it back
	state.QueryJson = plan.QueryJson

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *queryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.QueryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var detailedErr client.DetailedError
	query, err := r.client.Queries.Get(ctx, state.Dataset.ValueString(), state.ID.ValueString())
	if errors.As(err, &detailedErr) {
		if detailedErr.IsNotFound() {
			// if not found consider it deleted -- so just remove it from state
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.Append(helper.NewDetailedErrorDiagnostic(
				"Error Reading Honeycomb Query",
				&detailedErr,
			))
		}
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Honeycomb Query",
			"Could not read Query ID "+state.ID.ValueString()+": "+err.Error(),
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(*query.ID)
	// we don't encode the ID in the JSON
	query.ID = nil
	queryJson, err := query.Encode()
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Encoding Query", err) {
		return
	}
	state.QueryJson = types.StringValue(queryJson)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *queryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.QueryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Queries are immutable so just write the request's plan into the state's response
	// as described in the Migration Guide:
	//  https://developer.hashicorp.com/terraform/plugin/framework/migrating/resources/crud#migration-notes
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *queryResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Queries are immutable so we don't do anything but need to implement the method
	// to satisfy the interface
}
