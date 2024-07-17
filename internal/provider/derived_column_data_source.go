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
	_ datasource.DataSource              = &derivedColumnDataSource{}
	_ datasource.DataSourceWithConfigure = &derivedColumnDataSource{}
)

func NewDerivedColumnDataSource() datasource.DataSource {
	return &derivedColumnDataSource{}
}

// derivedColumnDataSource is the data source implementation.
type derivedColumnDataSource struct {
	client *client.Client
}

type derivedColumnDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Dataset     types.String `tfsdk:"dataset"`
	Alias       types.String `tfsdk:"alias"`
	Description types.String `tfsdk:"description"`
	Expression  types.String `tfsdk:"expression"`
}

func (d *derivedColumnDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_derived_column"
}

func (d *derivedColumnDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a Derived Column in a dataset",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"dataset": schema.StringAttribute{
				Description: "The dataset to fetch the derived column from. Use '__all__' to fetch an Environment-wide derived column.",
				Required:    true,
			},
			"alias": schema.StringAttribute{
				Description: "The alias of the Derived Column.",
				Required:    true,
			},
			"expression": schema.StringAttribute{
				Description: "The Derived Column's expression.",
				Computed:    true,
				Optional:    false,
				Required:    false,
			},
			"description": schema.StringAttribute{
				Description: "The Derived Column's description.",
				Computed:    true,
				Optional:    false,
				Required:    false,
			},
		},
	}
}

func (d *derivedColumnDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *derivedColumnDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data derivedColumnDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dc, err := d.client.DerivedColumns.GetByAlias(ctx, data.Dataset.ValueString(), data.Alias.ValueString())
	if helper.AddDiagnosticOnError(&resp.Diagnostics,
		fmt.Sprintf("Looking up Derived Column %q", data.ID.ValueString()),
		err) {
		return
	}

	data.ID = types.StringValue(dc.ID)
	data.Alias = types.StringValue(dc.Alias)
	data.Description = types.StringValue(dc.Description)
	data.Expression = types.StringValue(dc.Expression)

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
