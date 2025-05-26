package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
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
			"tags": schema.MapAttribute{
				Description: "A map of tags to assign to the resource.",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Map{
					modifiers.EquivalentTags(),
				},
				Validators: []validator.Map{
					mapvalidator.SizeAtMost(client.MaxTagsPerResource),
					mapvalidator.KeysAre(
						stringvalidator.RegexMatches(client.TagKeyValidationRegex, "must only contain lowercase letters, and be 1-32 characters long"),
					),
					mapvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(client.TagValueValidationRegex, "must begin with a lowercase letter, be between 1-32 characters long, and only contain lowercase alphanumeric characters, -, or /"),
					),
				},
			},
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
		Datasets: types.SetNull(types.StringType),
		Tags:     types.MapNull(types.StringType),
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

	// Convert the slice of strings to a types.Set
	if len(slo.DatasetSlugs) > 0 {
		ds := slo.DatasetSlugs
		if ds == nil {
			ds = []string{}
		}
		datasetsSet, diags := types.SetValueFrom(ctx, types.StringType, ds)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Datasets = datasetsSet
	} else {
		// Create an empty set if there are no dataset slugs
		state.Datasets = types.SetNull(types.StringType)
	}

	if len(slo.Tags) > 0 {
		tagMap := make(map[string]string)
		for _, tag := range slo.Tags {
			tagMap[tag.Key] = tag.Value
		}
		tags, diags := types.MapValueFrom(ctx, types.StringType, tagMap)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Tags = tags
	} else {
		// Create an empty map if there are no tags
		state.Tags = types.MapNull(types.StringType)
	}

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

	// Convert the slice of strings to a types.Set
	if len(slo.DatasetSlugs) > 0 {
		ds := slo.DatasetSlugs
		if ds == nil {
			ds = []string{}
		}
		datasetsSet, diags := types.SetValueFrom(ctx, types.StringType, ds)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Datasets = datasetsSet
	} else {
		// Create an empty set if there are no dataset slugs
		state.Datasets = types.SetNull(types.StringType)
	}

	if len(slo.Tags) > 0 {
		tagMap := make(map[string]string)
		for _, tag := range slo.Tags {
			tagMap[tag.Key] = tag.Value
		}
		tags, diags := types.MapValueFrom(ctx, types.StringType, tagMap)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Tags = tags
	} else {
		state.Tags = types.MapNull(types.StringType)
	}

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

	if len(slo.DatasetSlugs) > 0 {
		ds := slo.DatasetSlugs
		if ds == nil {
			ds = []string{}
		}
		datasetsSet, diags := types.SetValueFrom(ctx, types.StringType, ds)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Datasets = datasetsSet
	} else {
		// Create an empty set if there are no dataset slugs
		state.Datasets = types.SetNull(types.StringType)
	}

	if len(slo.Tags) > 0 {
		tagMap := make(map[string]string)
		for _, tag := range slo.Tags {
			tagMap[tag.Key] = tag.Value
		}
		tags, diags := types.MapValueFrom(ctx, types.StringType, tagMap)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Tags = tags
	} else {
		state.Tags = types.MapNull(types.StringType)
	}

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
	_, err := r.client.SLOs.Get(ctx, dataset.ValueString(), state.ID.ValueString())
	if errors.As(err, &detailedErr) && detailedErr.IsNotFound() {
		// Resource is already gone, no need to delete
		return
	}

	err = r.client.SLOs.Delete(ctx, dataset.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting SLO",
			fmt.Sprintf("Could not delete SLO %s: %s", state.ID.ValueString(), err),
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

	var tags []client.Tag
	if !plan.Tags.IsNull() {
		var tagMap map[string]string
		diags := plan.Tags.ElementsAs(ctx, &tagMap, false)
		if diags.HasError() {
			return nil, fmt.Errorf("error extracting tags: %v", diags)
		}
		for k, v := range tagMap {
			tags = append(tags, client.Tag{Key: k, Value: v})
		}
	} else {
		// if 'tags' is not present in the config, set to empty slice
		// to clear the tags
		tags = make([]client.Tag, 0)
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
