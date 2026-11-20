package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

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
	_ resource.Resource                = &datasetRestrictionResource{}
	_ resource.ResourceWithConfigure   = &datasetRestrictionResource{}
	_ resource.ResourceWithImportState = &datasetRestrictionResource{}
)

// datasetRestrictionResource restricts a dataset while it exists.
// Deleting it removes the restriction.
type datasetRestrictionResource struct {
	client *v2client.Client
}

func NewDatasetRestrictionResource() resource.Resource {
	return &datasetRestrictionResource{}
}

func (*datasetRestrictionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataset_restriction"
}

func (r *datasetRestrictionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (*datasetRestrictionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Restricts a Honeycomb Dataset. " +
			"Only Team owners and the members of a Group that has access to the Dataset can read a restricted Dataset.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the restriction, written as `<environment_id>/<dataset_id>`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Environment that holds the Dataset.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"dataset_id": schema.StringAttribute{
				Required: true,
				Description: "The unique ID of the Dataset to restrict. " +
					"It is the `hcid` of the `honeycombio_dataset` resource and data source, not the Dataset's slug.",
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

func (*datasetRestrictionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	environmentID, datasetID, found := strings.Cut(req.ID, "/")
	if !found || environmentID == "" || datasetID == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be written as <environment_id>/<dataset_id>",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &models.DatasetRestrictionResourceModel{
		ID:            types.StringValue(req.ID),
		EnvironmentID: types.StringValue(environmentID),
		DatasetID:     types.StringValue(datasetID),
	})...)
}

func (r *datasetRestrictionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.DatasetRestrictionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.EnvironmentID.ValueString() + "/" + plan.DatasetID.ValueString()
	ds, err := r.client.Datasets.Get(ctx, plan.EnvironmentID.ValueString(), plan.DatasetID.ValueString())
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.IsNotFound() {
		resp.Diagnostics.AddError(
			"Error Creating Honeycomb Dataset Restriction",
			fmt.Sprintf("The API found no Dataset %q in the Environment %q.",
				plan.DatasetID.ValueString(), plan.EnvironmentID.ValueString()),
		)
		return
	}
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Creating Honeycomb Dataset Restriction", err) {
		return
	}
	if ds.Restricted != nil && *ds.Restricted {
		resp.Diagnostics.AddError(
			"Honeycomb Dataset Is Already Restricted",
			"To manage the existing restriction with Terraform, import it, "+
				"for example `terraform import honeycombio_dataset_restriction.example "+id+"`.",
		)
		return
	}

	err = r.setRestricted(ctx, plan, true)
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Creating Honeycomb Dataset Restriction", err) {
		return
	}

	plan.ID = types.StringValue(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *datasetRestrictionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.DatasetRestrictionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ds, err := r.client.Datasets.Get(ctx, state.EnvironmentID.ValueString(), state.DatasetID.ValueString())
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.IsNotFound() {
		// the Dataset is gone, and its restriction with it
		resp.State.RemoveResource(ctx)
		return
	}
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Reading Honeycomb Dataset Restriction", err) {
		return
	}
	if ds.Restricted == nil || !*ds.Restricted {
		// the restriction was removed outside Terraform
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (*datasetRestrictionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.DatasetRestrictionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// every argument forces a replacement, so there is nothing to update
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *datasetRestrictionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.DatasetRestrictionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.setRestricted(ctx, state, false)
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.IsNotFound() {
		// the Dataset is already gone
		return
	}
	if !helper.AddDiagnosticOnError(&resp.Diagnostics, "Deleting Honeycomb Dataset Restriction", err) {
		return
	}
	if errors.As(err, &detailedErr) && detailedErr.Status == http.StatusForbidden {
		// the API's 403 does not say why
		resp.Diagnostics.AddError(
			"The Management Key cannot remove this Dataset Restriction",
			"If a Group has access to the Dataset, the key also needs the `groups:write` scope to remove the restriction.",
		)
	}
}

func (r *datasetRestrictionResource) setRestricted(
	ctx context.Context,
	m models.DatasetRestrictionResourceModel,
	restricted bool,
) error {
	_, err := r.client.Datasets.Update(ctx, m.EnvironmentID.ValueString(), &v2client.Dataset{
		ID:         m.DatasetID.ValueString(),
		Restricted: &restricted,
	})
	return err
}
