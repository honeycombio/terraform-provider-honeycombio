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

	groupID := plan.GroupID.ValueString()
	target := r.targetOf(plan)
	id := groupID + "/" + target.kind.name + "/" + target.id
	found, err := target.isGranted(ctx, groupID)
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

	err = target.kind.grant(ctx, groupID, []string{target.id})
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

	found, err := r.targetOf(state).isGranted(ctx, state.GroupID.ValueString())
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

	target := r.targetOf(state)
	err := target.kind.revoke(ctx, state.GroupID.ValueString(), []string{target.id})
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

// The names of the grant kinds, as written in the import ID.
const (
	datasetTarget     = "dataset"
	environmentTarget = "environment"
)

// grantKind is one kind of grant target, with the client calls for it.
type grantKind struct {
	name string // datasetTarget or environmentTarget
	// list returns the IDs of the Group's grants of this kind.
	list          func(ctx context.Context, groupID string, opts ...v2client.ListOption) ([]string, error)
	grant, revoke func(ctx context.Context, groupID string, ids []string) error
}

func datasetGrants(c *v2client.Client) grantKind {
	return grantKind{
		name: datasetTarget,
		list: func(ctx context.Context, groupID string, opts ...v2client.ListOption) ([]string, error) {
			pager, err := c.Groups.ListDatasets(ctx, groupID, append([]v2client.ListOption{v2client.PageSize(100)}, opts...)...)
			if err != nil {
				return nil, err
			}
			return pagerIDs(ctx, pager, func(d *v2client.GroupDataset) string { return d.ID })
		},
		grant:  c.Groups.GrantDatasets,
		revoke: c.Groups.RevokeDatasets,
	}
}

func environmentGrants(c *v2client.Client) grantKind {
	return grantKind{
		name: environmentTarget,
		list: func(ctx context.Context, groupID string, opts ...v2client.ListOption) ([]string, error) {
			pager, err := c.Groups.ListEnvironments(ctx, groupID, append([]v2client.ListOption{v2client.PageSize(100)}, opts...)...)
			if err != nil {
				return nil, err
			}
			return pagerIDs(ctx, pager, func(e *v2client.GroupEnvironment) string { return e.ID })
		},
		grant:  c.Groups.GrantEnvironments,
		revoke: c.Groups.RevokeEnvironments,
	}
}

// grantKinds are every kind of grant target.
func grantKinds(c *v2client.Client) []grantKind {
	return []grantKind{datasetGrants(c), environmentGrants(c)}
}

// grantTarget is what one grant gives access to.
type grantTarget struct {
	kind grantKind
	id   string
}

func (r *groupGrantResource) targetOf(m models.GroupGrantResourceModel) grantTarget {
	if !m.EnvironmentID.IsNull() {
		return grantTarget{kind: environmentGrants(r.client), id: m.EnvironmentID.ValueString()}
	}
	return grantTarget{kind: datasetGrants(r.client), id: m.DatasetID.ValueString()}
}

// idAttribute is the name of the argument that holds the target's ID.
func (t grantTarget) idAttribute() string {
	return t.kind.name + "_id"
}

// isGranted reports whether the API lists the target among the Group's grants.
//
// The filter makes the API return only the target, in one request. An older
// API ignores the filter and lists every grant. The ID check below still finds
// the target.
func (t grantTarget) isGranted(ctx context.Context, groupID string) (bool, error) {
	ids, err := t.kind.list(ctx, groupID, v2client.Filter("id", t.id))
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.Status == http.StatusUnprocessableEntity &&
		slices.ContainsFunc(detailedErr.Details, func(d hny.ErrorTypeDetail) bool { return d.Field == "parameter filter[id]" }) {
		// the API rejects a filter ID that is malformed or of the wrong
		// kind, and no grant has such an ID
		return false, nil
	}
	return slices.Contains(ids, t.id), err
}

// pagerIDs returns the ID of every item on every page of the pager.
func pagerIDs[T any](ctx context.Context, pager *v2client.Pager[T], idOf func(*T) string) ([]string, error) {
	ids := []string{}
	for pager.HasNext() {
		items, err := pager.Next(ctx)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			ids = append(ids, idOf(item))
		}
	}
	return ids, nil
}
