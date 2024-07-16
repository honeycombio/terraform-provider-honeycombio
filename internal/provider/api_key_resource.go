package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	hnyerr "github.com/honeycombio/terraform-provider-honeycombio/client/errors"
	v2client "github.com/honeycombio/terraform-provider-honeycombio/client/v2"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
//
// This resource is not implemeting ResourceWithImportState because importing keys
// won't give us the secret portion of the key which is arguably the whole reason
// for the resource.
var (
	_ resource.Resource              = &apiKeyResource{}
	_ resource.ResourceWithConfigure = &apiKeyResource{}
)

type apiKeyResource struct {
	client *v2client.Client
}

func NewAPIKeyResource() resource.Resource {
	return &apiKeyResource{}
}

func (*apiKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *apiKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (*apiKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "API keys are used to authenticate the Honeycomb API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this API key.",
				Computed:    true,
				Required:    false,
				Optional:    false,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the API Key.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The type of the API key.",
				Validators: []validator.String{
					stringvalidator.OneOf("ingest"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_id": schema.StringAttribute{
				Required:    true,
				Description: "The Environment ID the API key is scoped to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the API key is disabled.",
				Default:     booldefault.StaticBool(false),
			},
			"secret": schema.StringAttribute{
				Computed:    true,
				Required:    false,
				Optional:    false,
				Sensitive:   true,
				Description: "The secret portion of the API key. This is only available when creating a new key.",
			},
		},
		Blocks: map[string]schema.Block{
			"permissions": schema.ListNestedBlock{
				Description: "Permissions control what actions the API key can perform.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"create_datasets": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
							Description: "Allow this key to create missing datasets when sending telemetry.",
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *apiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.APIKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	newKey := &v2client.APIKey{
		Name:        plan.Name.ValueStringPointer(),
		KeyType:     plan.Type.ValueString(),
		Environment: &v2client.Environment{ID: plan.EnvironmentID.ValueString()},
		Disabled:    plan.Disabled.ValueBoolPointer(),
		Permissions: expandAPIKeyPermissions(ctx, plan.Permissions, &resp.Diagnostics),
	}

	key, err := r.client.APIKeys.Create(ctx, newKey)
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Creating Honeycomb API Key", err) {
		return
	}

	var state models.APIKeyResourceModel
	state.ID = types.StringValue(key.ID)
	state.Name = types.StringValue(*key.Name)
	state.Type = types.StringValue(key.KeyType)
	state.EnvironmentID = types.StringValue(key.Environment.ID)
	state.Disabled = types.BoolValue(*key.Disabled)
	state.Secret = types.StringValue(key.Secret)

	if !plan.Permissions.IsNull() {
		state.Permissions = flattenAPIKeyPermissions(ctx, key.Permissions, &resp.Diagnostics)
	} else {
		state.Permissions = types.ListNull(types.ObjectType{AttrTypes: models.APIKeyPermissionsAttrType})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *apiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.APIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var detailedErr hnyerr.DetailedError
	key, err := r.client.APIKeys.Get(ctx, state.ID.ValueString())
	if errors.As(err, &detailedErr) {
		if detailedErr.IsNotFound() {
			// if not found consider it deleted -- so just remove it from state
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.Append(helper.NewDetailedErrorDiagnostic(
				"Error Reading Honeycomb API Key",
				&detailedErr,
			))
		}
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Honeycomb API Key",
			"Unexpected error reading API Key ID "+state.ID.ValueString()+": "+err.Error(),
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(key.ID)
	state.Name = types.StringValue(*key.Name)
	state.Type = types.StringValue(key.KeyType)
	state.Disabled = types.BoolValue(*key.Disabled)
	state.EnvironmentID = types.StringValue(key.Environment.ID)

	if !state.Permissions.IsNull() {
		state.Permissions = flattenAPIKeyPermissions(ctx, key.Permissions, &resp.Diagnostics)
	} else {
		state.Permissions = types.ListNull(types.ObjectType{AttrTypes: models.APIKeyPermissionsAttrType})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *apiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state models.APIKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateRequest := &v2client.APIKey{
		ID:       plan.ID.ValueString(),
		Name:     plan.Name.ValueStringPointer(),
		Disabled: plan.Disabled.ValueBoolPointer(),
	}

	_, err := r.client.APIKeys.Update(ctx, updateRequest)
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Updating Honeycomb API Key", err) {
		return
	}

	key, err := r.client.APIKeys.Get(ctx, plan.ID.ValueString())
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Updating Honeycomb API Key", err) {
		return
	}

	state.ID = types.StringValue(key.ID)
	state.Name = types.StringValue(*key.Name)
	state.Type = types.StringValue(key.KeyType)
	state.Disabled = types.BoolValue(*key.Disabled)
	state.EnvironmentID = types.StringValue(key.Environment.ID)
	if !state.Permissions.IsNull() {
		state.Permissions = flattenAPIKeyPermissions(ctx, key.Permissions, &resp.Diagnostics)
	} else {
		state.Permissions = types.ListNull(types.ObjectType{AttrTypes: models.APIKeyPermissionsAttrType})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *apiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.APIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.APIKeys.Delete(ctx, state.ID.ValueString())
	var detailedErr hnyerr.DetailedError
	if err != nil {
		if errors.As(err, &detailedErr) {
			resp.Diagnostics.Append(helper.NewDetailedErrorDiagnostic(
				"Error Reading Honeycomb API Key",
				&detailedErr,
			))
		} else {
			resp.Diagnostics.AddError(
				"Error Deleting Honeycomb API Key",
				"Could not delete API Key ID "+state.ID.ValueString()+": "+err.Error(),
			)
		}
	}
}

func expandAPIKeyPermissions(ctx context.Context, list types.List, diags *diag.Diagnostics) *v2client.APIKeyPermissions {
	var permissions []models.APIKeyPermissionModel
	diags.Append(list.ElementsAs(ctx, &permissions, false)...)
	if diags.HasError() {
		return nil
	}

	if len(permissions) == 0 {
		return nil
	}

	return &v2client.APIKeyPermissions{
		CreateDatasets: permissions[0].CreateDatasets.ValueBool(),
	}
}

func flattenAPIKeyPermissions(ctx context.Context, p *v2client.APIKeyPermissions, diags *diag.Diagnostics) types.List {
	if p == nil {
		return types.ListNull(types.ObjectType{AttrTypes: models.APIKeyPermissionsAttrType})
	}

	obj, d := types.ObjectValue(models.APIKeyPermissionsAttrType, map[string]attr.Value{
		"create_datasets": types.BoolValue(p.CreateDatasets),
	})
	diags.Append(d...)

	result, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: models.APIKeyPermissionsAttrType}, []attr.Value{obj})
	diags.Append(d...)

	return result
}
