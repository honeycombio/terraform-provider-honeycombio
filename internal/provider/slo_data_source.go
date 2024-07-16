package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &sloDataSource{}
	_ datasource.DataSourceWithConfigure = &sloDataSource{}
)

func NewSLODataSource() datasource.DataSource {
	return &sloDataSource{}
}

// sloDataSource is the data source implementation.
type sloDataSource struct {
	client *client.Client
}

type sloDataSourceModel struct {
	ID               types.String  `tfsdk:"id"`
	Dataset          types.String  `tfsdk:"dataset"`
	Name             types.String  `tfsdk:"name"`
	Description      types.String  `tfsdk:"description"`
	SLI              types.String  `tfsdk:"sli"`
	TargetPercentage types.Float64 `tfsdk:"target_percentage"`
	TimePeriod       types.Int64   `tfsdk:"time_period"`
}

func (d *sloDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_slo"
}

func (d *sloDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches an SLO from a dataset",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the SLO to fetch.",
				Required:    true,
			},
			"dataset": schema.StringAttribute{
				Description: "The dataset to fetch the SLO from.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the SLO.",
				Computed:    true,
				Optional:    false,
				Required:    false,
			},
			"description": schema.StringAttribute{
				Description: "The SLO's description.",
				Computed:    true,
				Optional:    false,
				Required:    false,
			},
			"sli": schema.StringAttribute{
				Description: "The alias of the Derived Column used as the SLO's SLI.",
				Computed:    true,
				Optional:    false,
				Required:    false,
			},
			"target_percentage": schema.Float64Attribute{
				Description: "The percentage of qualified events expected to succeed during the `time_period`.",
				Computed:    true,
				Optional:    false,
				Required:    false,
			},
			"time_period": schema.Int64Attribute{
				Description: "The time period, in days, over which the SLO is evaluated.",
				Computed:    true,
				Optional:    false,
				Required:    false,
			},
		},
	}
}

func (d *sloDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sloDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data sloDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	slo, err := d.client.SLOs.Get(ctx, data.Dataset.ValueString(), data.ID.ValueString())
	if helper.AddDiagnosticOnError(&resp.Diagnostics,
		fmt.Sprintf("Looking up SLO %q", data.ID.ValueString()),
		err) {
		return
	}

	data.ID = types.StringValue(slo.ID)
	data.Name = types.StringValue(slo.Name)
	data.Description = types.StringValue(slo.Description)
	data.SLI = types.StringValue(slo.SLI.Alias)
	data.TargetPercentage = types.Float64Value(float64(slo.TargetPerMillion) / 10000)
	data.TimePeriod = types.Int64Value(int64(slo.TimePeriodDays))

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
