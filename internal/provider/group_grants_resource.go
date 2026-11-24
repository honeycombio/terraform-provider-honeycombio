package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
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
	_ resource.Resource                     = &groupGrantsResource{}
	_ resource.ResourceWithConfigure        = &groupGrantsResource{}
	_ resource.ResourceWithConfigValidators = &groupGrantsResource{}
	_ resource.ResourceWithImportState      = &groupGrantsResource{}
)

// groupGrantsResource owns every Dataset grant of a Group.
type groupGrantsResource struct {
	client *v2client.Client
}

func NewGroupGrantsResource() resource.Resource {
	return &groupGrantsResource{}
}

func (*groupGrantsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_grants"
}

func (r *groupGrantsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (*groupGrantsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	emptySet := types.SetValueMust(types.StringType, []attr.Value{})
	resp.Schema = schema.Schema{
		Description: "Manages all the grants of a Honeycomb Group, to Datasets and to Environments. Grants not listed are revoked.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Group whose grants to manage.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"dataset_ids": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(emptySet),
				ElementType: types.StringType,
				Description: "The unique IDs of every Dataset that the Group can read. " +
					"Each is the `hcid` of the `honeycombio_dataset` resource and data source, not the Dataset's slug. " +
					"Each Dataset must be restricted. Without this argument, the Group has no Dataset grants.",
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				},
			},
			"environment_ids": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				Default:     setdefault.StaticValue(emptySet),
				ElementType: types.StringType,
				Description: "The IDs of every Environment whose Datasets the Group can read. " +
					"Each is the `id` of the `honeycombio_environment` resource and data source, not the Environment's slug. " +
					"Without this argument, the Group has no Environment grants.",
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				},
			},
		},
	}
}

func (*groupGrantsResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	// a config with neither argument would revoke every grant, so ask for one
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(path.MatchRoot("dataset_ids"), path.MatchRoot("environment_ids")),
	}
}

func (*groupGrantsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		resp.Diagnostics.AddError("Invalid Import ID", "The Group ID must be provided")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &models.GroupGrantsModel{
		ID:             types.StringValue(req.ID),
		GroupID:        types.StringValue(req.ID),
		DatasetIDs:     types.SetNull(types.StringType),
		EnvironmentIDs: types.SetNull(types.StringType),
	})...)
}

func (r *groupGrantsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.GroupGrantsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := listGroupGrants(ctx, r.client, plan.GroupID.ValueString())
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.IsNotFound() {
		resp.Diagnostics.AddAttributeError(
			path.Root("group_id"),
			"Error Creating Honeycomb Group Grants",
			fmt.Sprintf("The API found no Group %q.", plan.GroupID.ValueString()),
		)
		return
	}
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Creating Honeycomb Group Grants", err) {
		return
	}
	if !current.isEmpty() {
		resp.Diagnostics.AddError(
			"Honeycomb Group Already Has Grants",
			"To manage the Group's existing grants with Terraform, import them, "+
				"for example `terraform import honeycombio_group_grants.example "+plan.GroupID.ValueString()+"`. "+
				"The next plan then shows how the grants differ from the configuration.",
		)
		return
	}

	resp.Diagnostics.Append(r.setGrants(ctx, "Creating Honeycomb Group Grants", plan, current)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = plan.GroupID
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *groupGrantsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.GroupGrantsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := listGroupGrants(ctx, r.client, state.GroupID.ValueString())
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.IsNotFound() {
		// the Group is gone, and its grants with it
		resp.State.RemoveResource(ctx)
		return
	}
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Reading Honeycomb Group Grants", err) {
		return
	}

	resp.Diagnostics.Append(current.storeIn(ctx, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *groupGrantsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.GroupGrantsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := listGroupGrants(ctx, r.client, plan.GroupID.ValueString())
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Updating Honeycomb Group Grants", err) {
		return
	}

	// a change of more than one batch can stop halfway; Read lists every
	// grant, so the next plan shows what is left
	resp.Diagnostics.Append(r.setGrants(ctx, "Updating Honeycomb Group Grants", plan, current)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *groupGrantsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.GroupGrantsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := state.GroupID.ValueString()
	current, err := listGroupGrants(ctx, r.client, groupID)
	for _, k := range grantKinds(r.client) {
		if err != nil {
			break
		}
		err = k.revoke(ctx, groupID, current[k.name])
	}
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.IsNotFound() {
		// the Group is already gone, and its grants with it
		return
	}
	helper.AddDiagnosticOnError(&resp.Diagnostics, "Deleting Honeycomb Group Grants", err)
}

// setGrants makes the Group's grants exactly the plan's dataset_ids and
// environment_ids. current is what the API lists now, so a target that was
// deleted is never sent: the API refuses a batch that names one.
func (r *groupGrantsResource) setGrants(ctx context.Context, summary string, plan models.GroupGrantsModel, current groupGrants) diag.Diagnostics {
	want, diags := grantsIn(ctx, plan)
	if diags.HasError() {
		return diags
	}

	groupID := plan.GroupID.ValueString()
	kinds := grantKinds(r.client)
	// grant everything first, so the members keep access to what stays
	for _, k := range kinds {
		if err := k.grant(ctx, groupID, subtract(want[k.name], current[k.name])); err != nil {
			return grantsError(summary, k, err)
		}
	}
	for _, k := range kinds {
		if err := k.revoke(ctx, groupID, subtract(current[k.name], want[k.name])); err != nil {
			return grantsError(summary, k, err)
		}
	}
	return nil
}

// grantsError turns an error from a grant or revoke of the kind into
// diagnostics. A 422 is about the IDs that the request sent, so it goes on
// the kind's attribute; the client's error names each ID the API rejected.
func grantsError(summary string, k grantKind, err error) diag.Diagnostics {
	var diags diag.Diagnostics
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.Status == http.StatusUnprocessableEntity {
		diags.AddAttributeError(path.Root(k.name+"_ids"), "Error "+summary, err.Error())
		return diags
	}
	helper.AddDiagnosticOnError(&diags, summary, err)
	return diags
}

// groupGrants are the IDs of a Group's grants, by the name of their kind.
type groupGrants map[string][]string

// listGroupGrants returns every grant of the Group.
func listGroupGrants(ctx context.Context, c *v2client.Client, groupID string) (groupGrants, error) {
	grants := groupGrants{}
	for _, k := range grantKinds(c) {
		ids, err := k.list(ctx, groupID)
		if err != nil {
			return nil, err
		}
		grants[k.name] = ids
	}
	return grants, nil
}

func (g groupGrants) isEmpty() bool {
	for _, ids := range g {
		if len(ids) > 0 {
			return false
		}
	}
	return true
}

// grantsIn returns the grants in the dataset_ids and environment_ids of m.
func grantsIn(ctx context.Context, m models.GroupGrantsModel) (groupGrants, diag.Diagnostics) {
	var datasetIDs, environmentIDs []string
	diags := m.DatasetIDs.ElementsAs(ctx, &datasetIDs, false)
	diags.Append(m.EnvironmentIDs.ElementsAs(ctx, &environmentIDs, false)...)
	return groupGrants{datasetTarget: datasetIDs, environmentTarget: environmentIDs}, diags
}

// storeIn stores the grants in the dataset_ids and environment_ids of m.
func (g groupGrants) storeIn(ctx context.Context, m *models.GroupGrantsModel) diag.Diagnostics {
	datasets, diags := types.SetValueFrom(ctx, types.StringType, g[datasetTarget])
	environments, d := types.SetValueFrom(ctx, types.StringType, g[environmentTarget])
	diags.Append(d...)
	m.DatasetIDs = datasets
	m.EnvironmentIDs = environments
	return diags
}

// subtract returns the items of a that are not in b.
func subtract(a, b []string) []string {
	var out []string
	for _, s := range a {
		if !slices.Contains(b, s) {
			out = append(out, s)
		}
	}
	return out
}
