package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/modifiers"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &sloResource{}
	_ resource.ResourceWithConfigure   = &sloResource{}
	_ resource.ResourceWithImportState = &sloResource{}
)

type sloResource struct {
	client *client.Client
}

func NewSLOResource() resource.Resource {
	return &sloResource{}
}

func (*sloResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_slo"
}

func (r *sloResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
}

func (*sloResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Honeycomb SLOs allows you to define and monitor Service Level Objectives (SLOs) for your organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the SLO.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the SLO.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 120),
				},
			},
			"description": schema.StringAttribute{
				Description: "A description of the SLO's intent and context.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1023),
				},
			},
			"dataset": schema.StringAttribute{
				Description: "The dataset this SLO is created in. Will be deprecated in a future release. Must be the same dataset as the SLI unless the SLI Derived Column is Environment-wide.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					modifiers.DatasetDeprecation(true),
				},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("datasets")),
				},
			},
			"datasets": schema.SetAttribute{
				Description: "The datasets the SLO is evaluated on.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.SizeAtMost(10),
					setvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("dataset")),
					setvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("dataset"), path.MatchRelative().AtParent().AtName("datasets")),
				},
			},
			"sli": schema.StringAttribute{
				Description: `The alias of the Derived Column that will be used as the SLI to indicate event success.
The derived column used as the SLI must be in the same dataset as the SLO. Additionally,
the column evaluation should consistently return nil, true, or false, as these are the only valid values for an SLI.`,
				Required: true,
			},
			"target_percentage": schema.Float64Attribute{
				Description: "The percentage of qualified events that you expect to succeed during the `time_period`.",
				Required:    true,
				Validators: []validator.Float64{
					float64validator.Between(0.00000, 99.9999),
				},
			},
			"time_period": schema.Int64Attribute{
				Description: "The time period, in days, over which your SLO will be evaluated.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func (r *sloResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	dataset, id, found := strings.Cut(req.ID, "/")

	// if dataset separator not found, we will assume its the bare id
	// if thats the case, we need to reassign values since strings.Cut would return (id, "", false)
	dsValue := types.StringNull()
	idValue := types.StringValue(id)
	if !found {
		idValue = types.StringValue(dataset)
	} else {
		dsValue = types.StringValue(dataset)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &models.SLOResourceModel{
		ID:       idValue,
		Dataset:  dsValue,
		Datasets: types.SetUnknown(types.StringType),
		Tags:     types.MapUnknown(types.StringType),
	})...)
}

func (r *sloResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan models.SLOResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataset := helper.GetDatasetOrAll(plan.Dataset)
	expandedSLO, err := expandSLO(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error expanding SLO for creation",
			fmt.Sprintf("Could not expand SLO: %s", err),
		)
		return
	}

	slo, err := r.client.SLOs.Create(ctx, dataset.ValueString(), expandedSLO)
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Creating Honeycomb SLO", err) {
		return
	}

	var state models.SLOResourceModel
	state.ID = types.StringValue(slo.ID)
	state.Name = types.StringValue(slo.Name)
	state.Description = types.StringValue(slo.Description)
	state.SLI = types.StringValue(slo.SLI.Alias)
	state.TargetPercentage = types.Float64Value(helper.PPMToFloat(slo.TargetPerMillion))
	state.TimePeriod = types.Int64Value(int64(slo.TimePeriodDays))

	if !plan.Dataset.IsNull() {
		state.Dataset = plan.Dataset
	}

	datasetsSet, diags := helper.DatasetSlugsToSet(ctx, slo.DatasetSlugs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Datasets = datasetsSet

	tags, diags := helper.TagsToMap(ctx, slo.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Tags = tags

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *sloResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state models.SLOResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataset := helper.GetDatasetOrAll(state.Dataset)

	var detailedErr client.DetailedError
	slo, err := r.client.SLOs.Get(ctx, dataset.ValueString(), state.ID.ValueString())
	if errors.As(err, &detailedErr) {
		if detailedErr.IsNotFound() {
			// if not found consider it deleted -- so just remove it from state
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.Append(helper.NewDetailedErrorDiagnostic(
			"Error Reading Honeycomb SLO",
			&detailedErr,
		))
		return
	} else if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Honeycomb SLO",
			fmt.Sprintf("Could not read SLO %s: %s", state.ID.ValueString(), err),
		)
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(slo.ID)
	state.Name = types.StringValue(slo.Name)
	state.Description = types.StringValue(slo.Description)
	state.SLI = types.StringValue(slo.SLI.Alias)
	state.TargetPercentage = types.Float64Value(helper.PPMToFloat(slo.TargetPerMillion))
	state.TimePeriod = types.Int64Value(int64(slo.TimePeriodDays))

	datasetsSet, diags := helper.DatasetSlugsToSet(ctx, slo.DatasetSlugs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Datasets = datasetsSet

	tags, diags := helper.TagsToMap(ctx, slo.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Tags = tags

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *sloResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan models.SLOResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataset := helper.GetDatasetOrAll(plan.Dataset)
	expandedSLO, err := expandSLO(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error expanding SLO for update",
			fmt.Sprintf("Could not expand SLO: %s", err),
		)
		return
	}

	slo, err := r.client.SLOs.Update(ctx, dataset.ValueString(), expandedSLO)
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Updating Honeycomb SLO", err) {
		return
	}

	var state models.SLOResourceModel
	state.ID = types.StringValue(slo.ID)
	state.Name = types.StringValue(slo.Name)
	state.Description = types.StringValue(slo.Description)
	state.SLI = types.StringValue(slo.SLI.Alias)
	state.TargetPercentage = types.Float64Value(helper.PPMToFloat(slo.TargetPerMillion))
	state.TimePeriod = types.Int64Value(int64(slo.TimePeriodDays))

	if !plan.Dataset.IsNull() {
		state.Dataset = plan.Dataset
	}

	datasetsSet, diags := helper.DatasetSlugsToSet(ctx, slo.DatasetSlugs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Datasets = datasetsSet

	tags, diags := helper.TagsToMap(ctx, slo.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Tags = tags

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *sloResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state models.SLOResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataset := helper.GetDatasetOrAll(state.Dataset)

	var detailedErr client.DetailedError
	err := r.client.SLOs.Delete(ctx, dataset.ValueString(), state.ID.ValueString())
	if err != nil {
		if errors.As(err, &detailedErr) {
			if detailedErr.IsNotFound() {
				return // if not found, consider it deleted
			}

			resp.Diagnostics.Append(helper.NewDetailedErrorDiagnostic(
				"Error deleting SLO",
				&detailedErr,
			))
			return
		}

		resp.Diagnostics.AddError(
			"Error deleting SLO",
			fmt.Sprintf("Could delete SLO %s: %s", state.ID.ValueString(), err),
		)
		return
	}
}

// expandSLO converts the plan to a SLO object for Honeycomb API.
// It handles the conversion of datasets and tags from the plan to the SLO object.
// If the "datasets" is not set, it uses the "dataset" as the DatasetSlugs.
// If the tags are not set, it uses an empty slice.
func expandSLO(ctx context.Context, plan models.SLOResourceModel) (*client.SLO, error) {
	var datasets []string
	if !plan.Datasets.IsNull() && !plan.Datasets.IsUnknown() {
		diags := plan.Datasets.ElementsAs(ctx, &datasets, false)
		if diags.HasError() {
			return nil, fmt.Errorf("error extracting datasets: %v", diags)
		}
	}

	tags, diags := helper.MapToTags(ctx, plan.Tags)
	if diags.HasError() {
		return nil, fmt.Errorf("error extracting tags: %v", diags)
	}

	slo := &client.SLO{
		ID:               plan.ID.ValueString(),
		Name:             plan.Name.ValueString(),
		Description:      plan.Description.ValueString(),
		TimePeriodDays:   int(plan.TimePeriod.ValueInt64()),
		TargetPerMillion: helper.FloatToPPM(plan.TargetPercentage.ValueFloat64()),
		SLI:              client.SLIRef{Alias: plan.SLI.ValueString()},
		DatasetSlugs:     datasets,
		Tags:             tags,
	}

	return slo, nil
}
