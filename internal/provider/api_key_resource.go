package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	v2client "github.com/honeycombio/terraform-provider-honeycombio/client/v2"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/validation"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
//
// This resource is not implementing ResourceWithImportState because importing keys
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
				MarkdownDescription: "The ID of the API Key.",
				Computed:            true,
				Required:            false,
				Optional:            false,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the API key.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
				},
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The type of API key. Currently only `ingest` and `configuration` is supported.",
				Validators: []validator.String{
					stringvalidator.OneOf("ingest", "configuration"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The Environment ID the API key is scoped to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"disabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the API key is disabled. Defaults to `false`.",
				Default:             booldefault.StaticBool(false),
			},
			"visible_to_members": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the key can be viewed by members and read-only users, or only owners.",
				Default:             booldefault.StaticBool(false),
			},
			"key": schema.StringAttribute{
				Computed:            true,
				Required:            false,
				Optional:            false,
				Sensitive:           true,
				MarkdownDescription: "The API key formatted for use based on its type.",
			},
			"secret": schema.StringAttribute{
				Computed:            true,
				Required:            false,
				Optional:            false,
				Sensitive:           true,
				MarkdownDescription: "The secret portion of the API Key.",
			},
		},
		Blocks: map[string]schema.Block{
			"permissions": schema.ListNestedBlock{
				MarkdownDescription: "A configuration block setting what actions the API key can perform.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"send_events": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
							MarkdownDescription: "Allow this configuration key to send events to Honeycomb. Defaults to `false`.",
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.RequiresReplace(),
							},
							Validators: []validator.Bool{
								boolvalidator.Any(
									validation.ParentValueValidator{
										Expression: path.MatchRoot("type"),
										Value:      types.DynamicValue(types.StringValue("configuration")),
									},
								),
							},
						},
						"create_datasets": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
							MarkdownDescription: "Allow this ingest or configuration key to create missing datasets when sending telemetry. Defaults to `false`.",
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.RequiresReplace(),
							},
						},
						"manage_queries": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
							MarkdownDescription: "Allow this configuration key to manage queries and columns. Defaults to `false`.",
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.RequiresReplace(),
							},
							Validators: []validator.Bool{
								validation.ParentValueValidator{
									Expression: path.MatchRoot("type"),
									Value:      types.DynamicValue(types.StringValue("configuration")),
								},
							},
						},
						"run_queries": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
							MarkdownDescription: "Allow this configuration key run queries. Defaults to `false`.",
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.RequiresReplace(),
							},
							Validators: []validator.Bool{
								validation.ParentValueValidator{
									Expression: path.MatchRoot("type"),
									Value:      types.DynamicValue(types.StringValue("configuration")),
								},
							},
						},
						"read_service_maps": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
							MarkdownDescription: "Allow this configuration key to read service maps. This feature is only for enterprise users. Defaults to `false`.",
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.RequiresReplace(),
							},
							Validators: []validator.Bool{
								validation.ParentValueValidator{
									Expression: path.MatchRoot("type"),
									Value:      types.DynamicValue(types.StringValue("configuration")),
								},
							},
						},
						"manage_public_boards": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
							MarkdownDescription: "Allow this configuration key to manage public boards. Defaults to `false`.",
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.RequiresReplace(),
							},
							Validators: []validator.Bool{
								validation.ParentValueValidator{
									Expression: path.MatchRoot("type"),
									Value:      types.DynamicValue(types.StringValue("configuration")),
								},
							},
						},
						"manage_private_boards": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
							MarkdownDescription: "Allow this configuration key to manage public boards. Defaults to `false`.",
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.RequiresReplace(),
							},
							Validators: []validator.Bool{
								validation.ParentValueValidator{
									Expression: path.MatchRoot("type"),
									Value:      types.DynamicValue(types.StringValue("configuration")),
								},
								validation.ParentValueValidator{
									Expression: path.MatchRoot("visible_to_members"),
									Value:      types.DynamicValue(types.BoolValue(false)),
								},
							},
						},
						"manage_slos": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
							MarkdownDescription: "Allow this configuration key to manage SLOs. Defaults to `false`.",
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.RequiresReplace(),
							},
							Validators: []validator.Bool{
								validation.ParentValueValidator{
									Expression: path.MatchRoot("type"),
									Value:      types.DynamicValue(types.StringValue("configuration")),
								},
							},
						},
						"manage_triggers": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
							MarkdownDescription: "Allow this configuration key to manage Triggers. Defaults to `false`.",
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.RequiresReplace(),
							},
							Validators: []validator.Bool{
								validation.ParentValueValidator{
									Expression: path.MatchRoot("type"),
									Value:      types.DynamicValue(types.StringValue("configuration")),
								},
							},
						},
						"manage_recipients": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
							MarkdownDescription: "Allow this configuration key to manage Recipients. Defaults to `false`.",
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.RequiresReplace(),
							},
							Validators: []validator.Bool{
								validation.ParentValueValidator{
									Expression: path.MatchRoot("type"),
									Value:      types.DynamicValue(types.StringValue("configuration")),
								},
							},
						},
						"manage_markers": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
							MarkdownDescription: "Allow this configuration key to manage Markers. Defaults to `false`.",
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.RequiresReplace(),
							},
							Validators: []validator.Bool{
								validation.ParentValueValidator{
									Expression: path.MatchRoot("type"),
									Value:      types.DynamicValue(types.StringValue("configuration")),
								},
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

	apiPermissions := expandAPIKeyPermissions(ctx, plan.Permissions, &resp.Diagnostics)
	if apiPermissions != nil {
		apiPermissions.VisibleToMembers = plan.VisibleToMembers.ValueBool()
	}

	newKey := &v2client.APIKey{
		Name:        plan.Name.ValueStringPointer(),
		KeyType:     plan.Type.ValueString(),
		Environment: &v2client.Environment{ID: plan.EnvironmentID.ValueString()},
		Disabled:    plan.Disabled.ValueBoolPointer(),
		Permissions: apiPermissions,
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
	state.VisibleToMembers = types.BoolValue(key.Permissions.VisibleToMembers)

	if !plan.Permissions.IsNull() {
		state.Permissions = flattenAPIKeyPermissions(ctx, key.Permissions, &resp.Diagnostics)
	} else {
		state.Permissions = types.ListNull(types.ObjectType{AttrTypes: models.APIKeyPermissionsAttrType})
	}

	switch key.KeyType {
	case "ingest":
		state.Key = types.StringValue(key.ID + key.Secret)
	case "configuration":
		state.Key = types.StringValue(key.Secret)
	default:
		resp.Diagnostics.AddError(
			"Unknown API Key Type",
			"API Key Type "+key.KeyType+" is not supported. Supported types are: ingest",
		)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *apiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.APIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var detailedErr client.DetailedError
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
	state.VisibleToMembers = types.BoolValue(key.Permissions.VisibleToMembers)

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
	state.VisibleToMembers = types.BoolValue(key.Permissions.VisibleToMembers)

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
	var detailedErr client.DetailedError
	if err != nil {
		if errors.As(err, &detailedErr) {
			// if not found consider it deleted -- so don't error
			if !detailedErr.IsNotFound() {
				resp.Diagnostics.Append(helper.NewDetailedErrorDiagnostic(
					"Error Deleting Honeycomb API Key",
					&detailedErr,
				))
			}
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
		SendEvents:          permissions[0].SendEvents.ValueBool(),
		CreateDatasets:      permissions[0].CreateDatasets.ValueBool(),
		ManageQueries:       permissions[0].ManageQueries.ValueBool(),
		RunQueries:          permissions[0].RunQueries.ValueBool(),
		ReadServiceMaps:     permissions[0].ReadServiceMaps.ValueBool(),
		ManagePublicBoards:  permissions[0].ManagePublicBoards.ValueBool(),
		ManagePrivateBoards: permissions[0].ManagePrivateBoards.ValueBool(),
		ManageSLOs:          permissions[0].ManageSLOs.ValueBool(),
		ManageTriggers:      permissions[0].ManageTriggers.ValueBool(),
		ManageRecipients:    permissions[0].ManageRecipients.ValueBool(),
		ManageMarkers:       permissions[0].ManageMarkers.ValueBool(),
	}
}

func flattenAPIKeyPermissions(ctx context.Context, p *v2client.APIKeyPermissions, diags *diag.Diagnostics) types.List {
	if p == nil {
		return types.ListNull(types.ObjectType{AttrTypes: models.APIKeyPermissionsAttrType})
	}

	obj, d := types.ObjectValue(models.APIKeyPermissionsAttrType, map[string]attr.Value{
		"send_events":           types.BoolValue(p.SendEvents),
		"create_datasets":       types.BoolValue(p.CreateDatasets),
		"manage_queries":        types.BoolValue(p.ManageQueries),
		"run_queries":           types.BoolValue(p.RunQueries),
		"read_service_maps":     types.BoolValue(p.ReadServiceMaps),
		"manage_public_boards":  types.BoolValue(p.ManagePublicBoards),
		"manage_private_boards": types.BoolValue(p.ManagePrivateBoards),
		"manage_slos":           types.BoolValue(p.ManageSLOs),
		"manage_triggers":       types.BoolValue(p.ManageTriggers),
		"manage_recipients":     types.BoolValue(p.ManageRecipients),
		"manage_markers":        types.BoolValue(p.ManageMarkers),
	})
	diags.Append(d...)

	result, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: models.APIKeyPermissionsAttrType}, []attr.Value{obj})
	diags.Append(d...)

	return result
}
