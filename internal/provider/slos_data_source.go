package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/hashcode"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &slosDataSource{}
	_ datasource.DataSourceWithConfigure = &slosDataSource{}
)

func NewSLOsDataSource() datasource.DataSource {
	return &slosDataSource{}
}

// slosDataSource is the data source implementation.
type slosDataSource struct {
	client *client.Client
}

func (d *slosDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_slos"
}

func (d *slosDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the SLOs in a dataset.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				Optional: false,
				Required: false,
			},
			"dataset": schema.StringAttribute{
				Description: "The dataset to fetch the SLOs from.",
				Optional:    true,
			},
			"ids": schema.ListAttribute{
				Description: "The list of SLO IDs.",
				Computed:    true,
				Optional:    false,
				Required:    false,
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"detail_filter": detailFilterSchema(),
		},
	}
}

func (d *slosDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	w := getClientFromDatasourceRequest(&req)
	if w == nil {
		return
	}

	c, err := w.V1Client()
	if err != nil || c == nil {
		resp.Diagnostics.AddError("Failed to configure client", err.Error())
		return
	}
	d.client = c
}

func (d *slosDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.SLOsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	datasetOrAll := helper.GetDatasetOrAll(data.Dataset)

	slos, err := d.client.SLOs.List(ctx, datasetOrAll.ValueString())
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Listing SLOs", err) {
		return
	}

	// Create a filter group with all filters (implicit AND logic)
	filterGroup, err := models.NewFilterGroup(data.DetailFilter)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create SLO filter group", err.Error())
		return
	}

	for _, s := range slos {
		sloResource, err := sloToResourceModel(ctx, &s)

		if helper.AddDiagnosticOnError(&resp.Diagnostics, "Converting SLO to resource model", err) {
			return
		}

		// Use the full filter matching capability for any field
		// For backward compatibility, if no filters are defined or if the filters
		// don't specifically target a field, include the SLO
		if filterGroup == nil || filterGroup.Match(sloResource) {
			data.IDs = append(data.IDs, types.StringValue(s.ID))
		}
	}
	data.ID = types.StringValue(hashcode.StringValues(data.IDs))

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// sloToResourceModel converts a client.SLO to a SLOResourceModel
func sloToResourceModel(ctx context.Context, slo *client.SLO) (*models.SLOResourceModel, error) {
	if slo == nil {
		return &models.SLOResourceModel{}, nil
	}

	datasetsSet, diags := helper.DatasetSlugsToSet(ctx, slo.DatasetSlugs)
	if diags.HasError() {
		return &models.SLOResourceModel{}, fmt.Errorf("error converting dataset slugs to set: %v", diags)
	}

	tags, diags := helper.TagsToMap(ctx, slo.Tags)
	if diags.HasError() {
		return &models.SLOResourceModel{}, fmt.Errorf("error converting tags to map: %v", diags)
	}

	return &models.SLOResourceModel{
		ID:               types.StringValue(slo.ID),
		Name:             types.StringValue(slo.Name),
		Description:      types.StringValue(slo.Description),
		SLI:              types.StringValue(slo.SLI.Alias),
		TargetPercentage: types.Float64Value(helper.PPMToFloat(slo.TargetPerMillion)),
		TimePeriod:       types.Int64Value(int64(slo.TimePeriodDays)),
		Datasets:         datasetsSet,
		Tags:             tags,
	}, nil
}
