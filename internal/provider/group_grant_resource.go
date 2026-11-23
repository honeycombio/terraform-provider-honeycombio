package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
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
	_ resource.Resource                     = &groupGrantResource{}
	_ resource.ResourceWithConfigure        = &groupGrantResource{}
	_ resource.ResourceWithConfigValidators = &groupGrantResource{}
	_ resource.ResourceWithImportState      = &groupGrantResource{}
)

// groupGrantResource gives a Group access to one target: a restricted
// Dataset, or every Dataset in an Environment.
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
		Description: "Gives a Honeycomb Group access to a restricted Dataset, or to every Dataset in an Environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				Description: "The ID of the grant, written as `<group_id>/dataset/<dataset_id>` " +
					"or `<group_id>/environment/<environment_id>`.",
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
				Optional: true,
				Description: "The unique ID of the Dataset that the Group can read. " +
					"It is the `hcid` of the `honeycombio_dataset` resource and data source, not the Dataset's slug. " +
					"The Dataset must be restricted. Set exactly one of `dataset_id` and `environment_id`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"environment_id": schema.StringAttribute{
				Optional: true,
				Description: "The ID of the Environment whose Datasets the Group can read, including Datasets added later. " +
					"It is the `id` of the `honeycombio_environment` resource and data source, not the Environment's slug. " +
					"The Environment does not have to be restricted. Set exactly one of `dataset_id` and `environment_id`.",
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

func (*groupGrantResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(path.MatchRoot("dataset_id"), path.MatchRoot("environment_id")),
	}
}

func (*groupGrantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 || parts[0] == "" || parts[2] == "" || (parts[1] != datasetTarget && parts[1] != environmentTarget) {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be written as <group_id>/dataset/<dataset_id> or <group_id>/environment/<environment_id>",
		)
		return
	}

	// the zero value of the other target's ID is null
	state := models.GroupGrantResourceModel{
		ID:      types.StringValue(req.ID),
		GroupID: types.StringValue(parts[0]),
	}
	if parts[1] == datasetTarget {
		state.DatasetID = types.StringValue(parts[2])
	} else {
		state.EnvironmentID = types.StringValue(parts[2])
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *groupGrantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.GroupGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	target := grantTargetOf(plan)
	id := plan.GroupID.ValueString() + "/" + target.kind + "/" + target.id
	found, err := r.hasGrant(ctx, plan.GroupID.ValueString(), target)
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

	err = r.grant(ctx, plan.GroupID.ValueString(), target)
	if errors.As(err, &detailedErr) && detailedErr.Status == http.StatusUnprocessableEntity {
		// the request holds only the target's ID, so a 422 is about it: the
		// target is unknown, or it is a Dataset that is not restricted
		resp.Diagnostics.AddAttributeError(path.Root(target.idAttribute()), "Error Creating Honeycomb Group Grant", err.Error())
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

	found, err := r.hasGrant(ctx, state.GroupID.ValueString(), grantTargetOf(state))
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
		// revoked outside Terraform, or removed by the API when the target
		// was deleted or, for a Dataset, lost its restriction
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

	err := r.revoke(ctx, state.GroupID.ValueString(), grantTargetOf(state))
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.IsNotFound() {
		// the Group is already gone
		return
	}
	if errors.As(err, &detailedErr) && detailedErr.Status == http.StatusUnprocessableEntity {
		// the API knows no target with this ID, so no grant is left
		return
	}
	helper.AddDiagnosticOnError(&resp.Diagnostics, "Deleting Honeycomb Group Grant", err)
}

// The kinds of grant target, as written in the import ID.
const (
	datasetTarget     = "dataset"
	environmentTarget = "environment"
)

// grantTarget is what a grant gives access to.
type grantTarget struct {
	kind string // datasetTarget or environmentTarget
	id   string
}

func grantTargetOf(m models.GroupGrantResourceModel) grantTarget {
	if !m.EnvironmentID.IsNull() {
		return grantTarget{kind: environmentTarget, id: m.EnvironmentID.ValueString()}
	}
	return grantTarget{kind: datasetTarget, id: m.DatasetID.ValueString()}
}

// idAttribute is the name of the argument that holds the target's ID.
func (t grantTarget) idAttribute() string {
	return t.kind + "_id"
}

func (r *groupGrantResource) grant(ctx context.Context, groupID string, t grantTarget) error {
	if t.kind == environmentTarget {
		return r.client.Groups.GrantEnvironments(ctx, groupID, []string{t.id})
	}
	return r.client.Groups.GrantDatasets(ctx, groupID, []string{t.id})
}

func (r *groupGrantResource) revoke(ctx context.Context, groupID string, t grantTarget) error {
	if t.kind == environmentTarget {
		return r.client.Groups.RevokeEnvironments(ctx, groupID, []string{t.id})
	}
	return r.client.Groups.RevokeDatasets(ctx, groupID, []string{t.id})
}

// hasGrant reports whether the API lists the target among the Group's grants.
func (r *groupGrantResource) hasGrant(ctx context.Context, groupID string, t grantTarget) (bool, error) {
	if t.kind == environmentTarget {
		pager, err := r.client.Groups.ListEnvironments(ctx, groupID, v2client.PageSize(100))
		if err != nil {
			return false, err
		}
		return pagerHasID(ctx, pager, t.id, func(e *v2client.GroupEnvironment) string { return e.ID })
	}
	pager, err := r.client.Groups.ListDatasets(ctx, groupID, v2client.PageSize(100))
	if err != nil {
		return false, err
	}
	return pagerHasID(ctx, pager, t.id, func(d *v2client.GroupDataset) string { return d.ID })
}

// pagerHasID reports whether an item on any page of the pager has the ID.
func pagerHasID[T any](ctx context.Context, pager *v2client.Pager[T], id string, idOf func(*T) string) (bool, error) {
	for pager.HasNext() {
		items, err := pager.Next(ctx)
		if err != nil {
			return false, err
		}
		if slices.ContainsFunc(items, func(item *T) bool { return idOf(item) == id }) {
			return true, nil
		}
	}
	return false, nil
}
