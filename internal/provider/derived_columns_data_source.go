package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
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
	IDs        []types.String `tfsdk:"names"`
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
				Optional:    false,
				Required:    false,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *derivedColumnsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *derivedColumnsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data derivedColumnsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	columns, err := d.client.DerivedColumns.List(ctx, data.Dataset.ValueString())
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Listing Derived Columns", err) {
		return
	}

	startsWith := data.StartsWith.ValueString()
	for _, s := range columns {
		if startsWith != "" && !strings.HasPrefix(s.Alias, startsWith) {
			continue
		}
		data.IDs = append(data.IDs, types.StringValue(s.ID))
	}
	data.ID = types.StringValue(hashcode.StringValues(data.IDs))

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
