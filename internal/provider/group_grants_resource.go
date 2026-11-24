package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	_ resource.Resource                = &groupGrantsResource{}
	_ resource.ResourceWithConfigure   = &groupGrantsResource{}
	_ resource.ResourceWithImportState = &groupGrantsResource{}
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
	resp.Schema = schema.Schema{
		Description: "Manages all the Dataset grants of a Honeycomb Group. Dataset grants not listed are revoked.",
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
				Required:    true,
				ElementType: types.StringType,
				Description: "The unique IDs of every Dataset that the Group can read. " +
					"Each is the `hcid` of the `honeycombio_dataset` resource and data source, not the Dataset's slug. " +
					"Each Dataset must be restricted. An empty set revokes every grant.",
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				},
			},
		},
	}
}

func (*groupGrantsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		resp.Diagnostics.AddError("Invalid Import ID", "The Group ID must be provided")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &models.GroupGrantsModel{
		ID:         types.StringValue(req.ID),
		GroupID:    types.StringValue(req.ID),
		DatasetIDs: types.SetNull(types.StringType),
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
	if len(current) > 0 {
		resp.Diagnostics.AddError(
			"Honeycomb Group Already Has Grants",
			"To manage the Group's existing grants with Terraform, import them, "+
				"for example `terraform import honeycombio_group_grants.example "+plan.GroupID.ValueString()+"`. "+
				"The next plan then shows how the grants differ from the configuration.",
		)
		return
	}

	if !r.setGrants(ctx, &resp.Diagnostics, "Creating Honeycomb Group Grants", plan, current) {
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

	datasetIDs, err := listGroupGrants(ctx, r.client, state.GroupID.ValueString())
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.IsNotFound() {
		// the Group is gone, and its grants with it
		resp.State.RemoveResource(ctx)
		return
	}
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Reading Honeycomb Group Grants", err) {
		return
	}

	set, diags := types.SetValueFrom(ctx, types.StringType, datasetIDs)
	resp.Diagnostics.Append(diags...)
	state.DatasetIDs = set
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
	if !r.setGrants(ctx, &resp.Diagnostics, "Updating Honeycomb Group Grants", plan, current) {
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
	if err == nil {
		err = r.client.Groups.RevokeDatasets(ctx, groupID, current)
	}
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.IsNotFound() {
		// the Group is already gone, and its grants with it
		return
	}
	helper.AddDiagnosticOnError(&resp.Diagnostics, "Deleting Honeycomb Group Grants", err)
}

// setGrants makes the Group's grants exactly the plan's dataset_ids.
// current is what the API lists now, so a Dataset that was deleted is
// never sent: the API refuses a batch that names one.
// It reports false when it added an error to diags.
func (r *groupGrantsResource) setGrants(
	ctx context.Context,
	diags *diag.Diagnostics,
	summary string,
	plan models.GroupGrantsModel,
	current []string,
) bool {
	var want []string
	diags.Append(plan.DatasetIDs.ElementsAs(ctx, &want, false)...)
	if diags.HasError() {
		return false
	}

	groupID := plan.GroupID.ValueString()
	toGrant := subtract(want, current)
	toRevoke := subtract(current, want)
	// grant first, so the members keep access to the Datasets that stay
	err := r.client.Groups.GrantDatasets(ctx, groupID, toGrant)
	if err == nil {
		err = r.client.Groups.RevokeDatasets(ctx, groupID, toRevoke)
	}
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.Status == http.StatusUnprocessableEntity {
		// the requests hold only dataset IDs, so a 422 is about them;
		// the client's error names each dataset ID the API rejected
		diags.AddAttributeError(path.Root("dataset_ids"), "Error "+summary, err.Error())
		return false
	}
	return !helper.AddDiagnosticOnError(diags, summary, err)
}

// listGroupGrants returns the IDs of every Dataset the Group is granted.
func listGroupGrants(ctx context.Context, c *v2client.Client, groupID string) ([]string, error) {
	pager, err := c.Groups.ListDatasets(ctx, groupID, v2client.PageSize(100))
	if err != nil {
		return nil, err
	}

	datasetIDs := []string{}
	for pager.HasNext() {
		datasets, err := pager.Next(ctx)
		if err != nil {
			return nil, err
		}
		for _, ds := range datasets {
			datasetIDs = append(datasetIDs, ds.ID)
		}
	}
	return datasetIDs, nil
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
