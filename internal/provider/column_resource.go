package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &columnResource{}
	_ resource.ResourceWithConfigure   = &columnResource{}
	_ resource.ResourceWithImportState = &columnResource{}
)

type columnResource struct {
	client *client.Client
}

func NewColumnResource() resource.Resource {
	return &columnResource{}
}

func (*columnResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_column"
}

func (r *columnResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	w := getClientFromResourceRequest(&req)
	if w == nil {
		return
	}

	c, err := w.V1Client()
	if err != nil || c == nil {
		resp.Diagnostics.AddError("Unable to create client", err.Error())
		return
	}
	r.client = c
}

func (*columnResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Honeycomb Column resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Column.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Column.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"dataset": schema.StringAttribute{
				MarkdownDescription: "The dataset this Column belongs to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hidden": schema.BoolAttribute{
				MarkdownDescription: "Whether the Column is hidden or not.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The Column's description.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The Column's type. Valid values are `string`, `integer`, `float`, `boolean`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("string"),
				Validators: []validator.String{
					stringvalidator.OneOf(helper.AsStringSlice(client.ColumnTypes())...),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The time the Column was created.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "The time the Column was last updated.",
				Computed:            true,
			},
			"last_written_at": schema.StringAttribute{
				MarkdownDescription: "The time the Column was last written to.",
				Computed:            true,
			},
		},
	}
}

func (r *columnResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.ColumnResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	column, err := r.client.Columns.Create(ctx, plan.Dataset.ValueString(), r.expandColumn(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error creating Column", err.Error())
		return
	}

	r.updateModelFromColumn(&plan, column)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *columnResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.ColumnResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	columnName := state.Name.ValueString()

	column, err := r.client.Columns.GetByKeyName(ctx, state.Dataset.ValueString(), columnName)
	if err != nil {
		var detailedError client.DetailedError
		if errors.As(err, &detailedError) && detailedError.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading Column %s", columnName),
			err.Error(),
		)
		return
	}

	r.updateModelFromColumn(&state, column)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *columnResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.ColumnResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	column, err := r.client.Columns.Update(ctx, plan.Dataset.ValueString(), r.expandColumn(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating Column", err.Error())
		return
	}

	r.updateModelFromColumn(&plan, column)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *columnResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.ColumnResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Columns.Delete(ctx, state.Dataset.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting Column", err.Error())
		return
	}
}

func (r *columnResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// import ID is of the format <dataset>/<column name>
	dataset, name, found := strings.Cut(req.ID, "/")
	if !found {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Import ID must be written as <dataset>/<column name>",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &models.ColumnResourceModel{
		Name:    types.StringValue(name),
		Dataset: types.StringValue(dataset),
	})...)
}

func (r *columnResource) expandColumn(model models.ColumnResourceModel) *client.Column {
	columnName := model.Name.ValueString()

	column := &client.Column{
		ID:      model.ID.ValueString(),
		KeyName: columnName,
	}

	if !model.Hidden.IsNull() {
		hidden := model.Hidden.ValueBool()
		column.Hidden = &hidden
	}

	if !model.Description.IsNull() {
		column.Description = model.Description.ValueString()
	}

	if !model.Type.IsNull() {
		columnType := client.ColumnType(model.Type.ValueString())
		column.Type = &columnType
	}

	return column
}

func (r *columnResource) updateModelFromColumn(model *models.ColumnResourceModel, column *client.Column) {
	model.ID = types.StringValue(column.ID)
	model.Name = types.StringValue(column.KeyName)

	if column.Hidden != nil {
		model.Hidden = types.BoolValue(*column.Hidden)
	} else {
		model.Hidden = types.BoolValue(false)
	}

	model.Description = types.StringValue(column.Description)

	if column.Type != nil {
		model.Type = types.StringValue(string(*column.Type))
	} else {
		model.Type = types.StringValue("string")
	}

	if !column.CreatedAt.IsZero() {
		model.CreatedAt = types.StringValue(column.CreatedAt.UTC().Format(time.RFC3339))
	}

	if !column.UpdatedAt.IsZero() {
		model.UpdatedAt = types.StringValue(column.UpdatedAt.UTC().Format(time.RFC3339))
	}

	if !column.LastWrittenAt.IsZero() {
		model.LastWrittenAt = types.StringValue(column.LastWrittenAt.UTC().Format(time.RFC3339))
	}
}
