package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/hashcode"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &derivedColumnsDataSource{}
	_ datasource.DataSourceWithConfigure = &derivedColumnsDataSource{}
)

func NewDerivedColumnsDataSource() datasource.DataSource {
	return &derivedColumnsDataSource{}
}

// derivedColumnsDataSource is the data source implementation.
type derivedColumnsDataSource struct {
	client *client.Client
}

type derivedColumnsDataSourceModel struct {
	ID         types.String   `tfsdk:"id"`
	Dataset    types.String   `tfsdk:"dataset"`
	StartsWith types.String   `tfsdk:"starts_with"`
	Columns    []types.String `tfsdk:"names"`
}

func (d *derivedColumnsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_derived_columns"
}

func (d *derivedColumnsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
			"starts_with": schema.StringAttribute{
				Description: "Only return columns starting with the given value.",
				Optional:    true,
			},
			"names": schema.ListAttribute{
				Description: "The list of Derived Column names.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *derivedColumnsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	d.client = getClientFromDatasourceRequest(&req)
}

func (d *derivedColumnsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data derivedColumnsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	columns, err := d.client.DerivedColumns.List(ctx, data.Dataset.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Derived Columns", err.Error())
		return
	}

	if !data.StartsWith.IsNull() {
		startsWith := data.StartsWith.ValueString()

		for i := len(columns) - 1; i >= 0; i-- {
			if !strings.HasPrefix(columns[i].Alias, startsWith) {
				columns = append(columns[:i], columns[i+1:]...)
			}
		}
	}

	ids := make([]string, len(columns))
	for _, dc := range columns {
		data.Columns = append(data.Columns, types.StringValue(dc.Alias))
	}
	data.ID = types.StringValue(hashcode.Strings(ids))

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
