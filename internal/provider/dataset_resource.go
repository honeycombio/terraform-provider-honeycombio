package provider

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/features"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/modifiers"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &datasetResource{}
	_ resource.ResourceWithConfigure   = &datasetResource{}
	_ resource.ResourceWithModifyPlan  = &datasetResource{}
	_ resource.ResourceWithImportState = &datasetResource{}
)

type datasetResource struct {
	client  *client.Client
	feature features.FeaturesDataset
}

func NewDatasetResource() resource.Resource {
	return &datasetResource{}
}

func (*datasetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataset"
}

func (r *datasetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	features, err := w.Features()
	if err != nil {
		resp.Diagnostics.AddError("Unable to get features", err.Error())
		return
	}
	r.feature = features.Dataset
}

func (*datasetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dataset in a Honeycomb Environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Required:    false,
				Optional:    false,
				Description: "The ID of the Dataset.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Dataset.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"slug": schema.StringAttribute{
				Description: "The slug of the Dataset.",
				Computed:    true,
				Optional:    false,
				Required:    false,
			},
			"description": schema.StringAttribute{
				Description: "The Dataset's description.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.LengthAtMost(1024),
				},
			},
			"expand_json_depth": schema.Int32Attribute{
				Description: "The maximum unpacking depth of nested JSON fields.",
				Computed:    true,
				Optional:    true,
				Default:     int32default.StaticInt32(0),
				Validators: []validator.Int32{
					int32validator.Between(0, 10),
				},
			},
			"delete_protected": schema.BoolAttribute{
				Description: "The current delete protection status of the Dataset. Cannot be set to false on creation.",
				Default:     booldefault.StaticBool(true),
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					modifiers.EnforceDeletionProtection(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "ISO8601-formatted time the dataset was created.",
				Computed:    true,
				Required:    false,
				Optional:    false,
			},
			"last_written_at": schema.StringAttribute{
				Description: "ISO8601-formatted time the dataset was last written to (received event data).",
				Computed:    true,
				Required:    false,
				Optional:    false,
			},
		},
	}
}

func (r *datasetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID == "" {
		resp.Diagnostics.AddError("Invalid Import ID", "The Dataset Slug must be provided")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &models.DatasetResourceModel{
		ID: types.StringValue(req.ID),
	})...)
}

func (r *datasetResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		// If the entire plan is null, the resource is planned for destruction -- let's add a warning
		resp.Diagnostics.AddWarning(
			"Resource Destruction Warning",
			"Applying this plan will delete the Dataset and all of its contents. "+
				"This is an irreversible operation.",
		)
	}
}

func (r *datasetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.DatasetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ds, err := r.client.Datasets.Create(ctx, &client.Dataset{
		Name:            plan.Name.ValueString(),
		Description:     plan.Description.ValueString(),
		ExpandJSONDepth: int(plan.ExpandJSONDepth.ValueInt32()),
	})
	if errors.Is(err, client.ErrDatasetExists) {
		if r.feature.ImportOnConflict {
			// if the dataset already exists and import_on_conflict is true,
			// we should import the existing dataset instead of creating a new one.
			resp.Diagnostics.AddWarning("Importing existing Dataset on Create",
				"Dataset \""+plan.Name.ValueString()+"\" already exists, importing and updating the existing dataset as "+
					"'import_on_conflict' is enabled.")

			// read the existing dataset back
			ds, err = r.client.Datasets.GetByName(ctx, plan.Name.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Error importing existing Dataset", err.Error())
				return
			}

			// update the dataset with the plan's values
			ds, err = r.client.Datasets.Update(ctx, &client.Dataset{
				Slug:            ds.Slug,
				Description:     plan.Description.ValueString(),
				ExpandJSONDepth: int(plan.ExpandJSONDepth.ValueInt32()),
			})
			if err != nil {
				resp.Diagnostics.AddError("Error updating imported Dataset", err.Error())
				return
			}
		} else {
			resp.Diagnostics.AddError("Error Creating Honeycomb Dataset", "Dataset \""+plan.Name.ValueString()+"\" already exists")
			return
		}
	} else if helper.AddDiagnosticOnError(&resp.Diagnostics, "Creating Dataset", err) {
		return
	}

	var state models.DatasetResourceModel
	state.ID = types.StringValue(ds.Slug)
	state.Slug = types.StringValue(ds.Slug)
	state.Name = types.StringValue(ds.Name)
	state.Description = types.StringValue(ds.Description)
	state.ExpandJSONDepth = types.Int32Value(int32(ds.ExpandJSONDepth))
	state.DeleteProtected = types.BoolPointerValue(ds.Settings.DeleteProtected)
	state.CreatedAt = types.StringValue(ds.CreatedAt.Format(time.RFC3339))
	state.LastWrittenAt = types.StringValue(ds.LastWrittenAt.Format(time.RFC3339))

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *datasetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.DatasetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var detailedErr client.DetailedError
	ds, err := r.client.Datasets.Get(ctx, state.ID.ValueString())
	if errors.As(err, &detailedErr) {
		if detailedErr.IsNotFound() {
			// if not found consider it deleted -- so just remove it from state
			resp.State.RemoveResource(ctx)
			return
		} else {
			resp.Diagnostics.Append(helper.NewDetailedErrorDiagnostic(
				"Error Reading Honeycomb Dataset",
				&detailedErr,
			))
		}
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Honeycomb Dataset",
			"Unexpected error reading Dataset "+state.ID.ValueString()+": "+err.Error(),
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(ds.Slug)
	state.Name = types.StringValue(ds.Name)
	state.Slug = types.StringValue(ds.Slug)
	state.Description = types.StringValue(ds.Description)
	state.ExpandJSONDepth = types.Int32Value(int32(ds.ExpandJSONDepth))
	state.DeleteProtected = types.BoolPointerValue(ds.Settings.DeleteProtected)
	state.CreatedAt = types.StringValue(ds.CreatedAt.Format(time.RFC3339))
	state.LastWrittenAt = types.StringValue(ds.LastWrittenAt.Format(time.RFC3339))

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *datasetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state models.DatasetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Datasets.Update(ctx, &client.Dataset{
		Slug:            plan.ID.ValueString(),
		Description:     plan.Description.ValueString(),
		ExpandJSONDepth: int(plan.ExpandJSONDepth.ValueInt32()),
		Settings: client.DatasetSettings{
			DeleteProtected: plan.DeleteProtected.ValueBoolPointer(),
		},
	})
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Updating Honeycomb Dataset", err) {
		return
	}

	ds, err := r.client.Datasets.Get(ctx, plan.ID.ValueString())
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Updating Honeycomb Dataset", err) {
		return
	}

	state.Name = types.StringValue(ds.Name)
	state.Slug = types.StringValue(ds.Slug)
	state.Description = types.StringValue(ds.Description)
	state.ExpandJSONDepth = types.Int32Value(int32(ds.ExpandJSONDepth))
	state.DeleteProtected = types.BoolPointerValue(ds.Settings.DeleteProtected)
	state.CreatedAt = types.StringValue(ds.CreatedAt.Format(time.RFC3339))
	state.LastWrittenAt = types.StringValue(ds.LastWrittenAt.Format(time.RFC3339))

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *datasetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.DatasetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Datasets.Delete(ctx, state.ID.ValueString())
	var detailedErr client.DetailedError
	if err != nil {
		if errors.As(err, &detailedErr) {
			if detailedErr.Status == http.StatusConflict {
				resp.Diagnostics.AddError(
					"Unable to Delete Dataset",
					"Delete Protection is enabled. "+
						"You must disable delete protection before it can be deleted.",
				)
			} else {
				resp.Diagnostics.Append(helper.NewDetailedErrorDiagnostic(
					"Error Deleting Honeycomb Dataset",
					&detailedErr,
				))
			}
		} else {
			resp.Diagnostics.AddError(
				"Error Deleting Honeycomb Dataset",
				"Could not delete Dataset "+state.ID.ValueString()+": "+err.Error(),
			)
		}
	}
}
