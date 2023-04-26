package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
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
		Description: "Fetches the Derived Columns in a dataset",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"dataset": schema.StringAttribute{
				Description: "The dataset to fetch the derived columns from. Use '__all__' to fetch Environment-wide derived columns.",
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

func (d *derivedColumnDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	d.client = getClientFromDatasourceRequest(&req)
}

func (d *derivedColumnDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data derivedColumnDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dc, err := d.client.DerivedColumns.GetByAlias(ctx, data.Dataset.ValueString(), data.Alias.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to lookup Derived Column \"%s\"", data.Alias.ValueString()),
			err.Error())
		return
	}

	data.ID = types.StringValue(dc.ID)
	data.Alias = types.StringValue(dc.Alias)
	data.Description = types.StringValue(dc.Description)
	data.Expression = types.StringValue(dc.Expression)

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
