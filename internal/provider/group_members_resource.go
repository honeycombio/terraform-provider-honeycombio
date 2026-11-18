package provider

import (
	"context"
	"errors"
	"net/http"

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
	_ resource.Resource                = &groupMembersResource{}
	_ resource.ResourceWithConfigure   = &groupMembersResource{}
	_ resource.ResourceWithImportState = &groupMembersResource{}
)

// maxGroupMembers is the most users the API accepts in one replace request.
const maxGroupMembers = 15000

type groupMembersResource struct {
	client *v2client.Client
}

func NewGroupMembersResource() resource.Resource {
	return &groupMembersResource{}
}

func (*groupMembersResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_members"
}

func (r *groupMembersResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (*groupMembersResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages all the users in a Honeycomb Group. Users not listed are removed from the Group.",
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
				Description: "The ID of the Group whose users to manage.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"user_ids": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "The IDs of every user in the Group. An empty set removes every user.",
				Validators: []validator.Set{
					setvalidator.SizeAtMost(maxGroupMembers),
					setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				},
			},
		},
	}
}

func (r *groupMembersResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		resp.Diagnostics.AddError("Invalid Import ID", "The Group ID must be provided")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &models.GroupMembersResourceModel{
		ID:      types.StringValue(req.ID),
		GroupID: types.StringValue(req.ID),
		UserIDs: types.SetNull(types.StringType),
	})...)
}

func (r *groupMembersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.GroupMembersResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !r.replaceMembers(ctx, &resp.Diagnostics, "Creating Honeycomb Group Members", plan) {
		return
	}

	plan.ID = plan.GroupID
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *groupMembersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.GroupMembersResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userIDs, err := r.listMembers(ctx, state.GroupID.ValueString())
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.IsNotFound() {
		// the Group is gone, and its members with it
		resp.State.RemoveResource(ctx)
		return
	}
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Reading Honeycomb Group Members", err) {
		return
	}

	set, diags := types.SetValueFrom(ctx, types.StringType, userIDs)
	resp.Diagnostics.Append(diags...)
	state.UserIDs = set
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *groupMembersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.GroupMembersResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !r.replaceMembers(ctx, &resp.Diagnostics, "Updating Honeycomb Group Members", plan) {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *groupMembersResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.GroupMembersResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// remove every user; agent members stay
	err := r.client.Groups.ReplaceMembers(ctx, state.GroupID.ValueString(), []string{})
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.IsNotFound() {
		// the Group is already gone
		return
	}
	helper.AddDiagnosticOnError(&resp.Diagnostics, "Deleting Honeycomb Group Members", err)
}

// replaceMembers makes the Group's users exactly the plan's user_ids.
// It reports false when it added an error to diags.
func (r *groupMembersResource) replaceMembers(
	ctx context.Context,
	diags *diag.Diagnostics,
	summary string,
	plan models.GroupMembersResourceModel,
) bool {
	var userIDs []string
	diags.Append(plan.UserIDs.ElementsAs(ctx, &userIDs, false)...)
	if diags.HasError() {
		return false
	}

	err := r.client.Groups.ReplaceMembers(ctx, plan.GroupID.ValueString(), userIDs)
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.Status == http.StatusUnprocessableEntity {
		// the request holds only user_ids, so a 422 is about them;
		// the client's error names each user ID the API rejected
		diags.AddAttributeError(path.Root("user_ids"), "Error "+summary, err.Error())
		return false
	}
	return !helper.AddDiagnosticOnError(diags, summary, err)
}

// listMembers returns the IDs of every user in the Group.
func (r *groupMembersResource) listMembers(ctx context.Context, groupID string) ([]string, error) {
	pager, err := r.client.Groups.ListMembers(ctx, groupID, v2client.PageSize(100))
	if err != nil {
		return nil, err
	}

	userIDs := []string{}
	for pager.HasNext() {
		members, err := pager.Next(ctx)
		if err != nil {
			return nil, err
		}
		for _, m := range members {
			userIDs = append(userIDs, m.ID)
		}
	}
	return userIDs, nil
}
