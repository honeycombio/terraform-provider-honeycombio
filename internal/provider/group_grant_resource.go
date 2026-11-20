package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	_ resource.Resource                = &groupGrantResource{}
	_ resource.ResourceWithConfigure   = &groupGrantResource{}
	_ resource.ResourceWithImportState = &groupGrantResource{}
)

// groupGrantResource gives a Group access to one restricted Dataset.
type groupGrantResource struct {
	client *v2client.Client
}

func NewGroupGrantResource() resource.Resource {
	return &groupGrantResource{}
}

func (*groupGrantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_grant"
}

func (r *groupGrantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (*groupGrantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Gives a Honeycomb Group access to a restricted Dataset. The members of the Group can read the Dataset.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the grant, written as `<group_id>/dataset/<dataset_id>`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Group to give access to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"dataset_id": schema.StringAttribute{
				Required: true,
				Description: "The unique ID of the Dataset that the Group can read. " +
					"It is the `hcid` of the `honeycombio_dataset` resource and data source, not the Dataset's slug. " +
					"The Dataset must be restricted.",
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

func (*groupGrantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 || parts[0] == "" || parts[1] != "dataset" || parts[2] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be written as <group_id>/dataset/<dataset_id>",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &models.GroupGrantResourceModel{
		ID:        types.StringValue(req.ID),
		GroupID:   types.StringValue(parts[0]),
		DatasetID: types.StringValue(parts[2]),
	})...)
}

func (r *groupGrantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.GroupGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.GroupID.ValueString() + "/dataset/" + plan.DatasetID.ValueString()
	found, err := r.hasGrant(ctx, plan)
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.IsNotFound() {
		resp.Diagnostics.AddAttributeError(
			path.Root("group_id"),
			"Error Creating Honeycomb Group Grant",
			fmt.Sprintf("The API found no Group %q.", plan.GroupID.ValueString()),
		)
		return
	}
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Creating Honeycomb Group Grant", err) {
		return
	}
	if found {
		resp.Diagnostics.AddError(
			"Honeycomb Group Grant Already Exists",
			"To manage the existing grant with Terraform, import it, "+
				"for example `terraform import honeycombio_group_grant.example "+id+"`.",
		)
		return
	}

	err = r.client.Groups.GrantDatasets(ctx, plan.GroupID.ValueString(), []string{plan.DatasetID.ValueString()})
	if errors.As(err, &detailedErr) && detailedErr.Status == http.StatusUnprocessableEntity {
		// the request holds only the dataset ID, so a 422 is about it:
		// the Dataset is unknown or not restricted
		resp.Diagnostics.AddAttributeError(path.Root("dataset_id"), "Error Creating Honeycomb Group Grant", err.Error())
		return
	}
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Creating Honeycomb Group Grant", err) {
		return
	}

	plan.ID = types.StringValue(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *groupGrantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.GroupGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, err := r.hasGrant(ctx, state)
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.IsNotFound() {
		// the Group is gone, and its grants with it
		resp.State.RemoveResource(ctx)
		return
	}
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Reading Honeycomb Group Grant", err) {
		return
	}
	if !found {
		// revoked outside Terraform, or removed by the API when the
		// Dataset lost its restriction or was deleted
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (*groupGrantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.GroupGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// every argument forces a replacement, so there is nothing to update
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *groupGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.GroupGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Groups.RevokeDatasets(ctx, state.GroupID.ValueString(), []string{state.DatasetID.ValueString()})
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.IsNotFound() {
		// the Group is already gone
		return
	}
	if errors.As(err, &detailedErr) && detailedErr.Status == http.StatusUnprocessableEntity {
		// the API knows no Dataset with this ID, so no grant is left
		return
	}
	helper.AddDiagnosticOnError(&resp.Diagnostics, "Deleting Honeycomb Group Grant", err)
}

// hasGrant reports whether the API lists the grant's Dataset for its Group.
func (r *groupGrantResource) hasGrant(ctx context.Context, m models.GroupGrantResourceModel) (bool, error) {
	pager, err := r.client.Groups.ListDatasets(ctx, m.GroupID.ValueString(), v2client.PageSize(100))
	if err != nil {
		return false, err
	}
	for pager.HasNext() {
		datasets, err := pager.Next(ctx)
		if err != nil {
			return false, err
		}
		for _, ds := range datasets {
			if ds.ID == m.DatasetID.ValueString() {
				return true, nil
			}
		}
	}
	return false, nil
}
