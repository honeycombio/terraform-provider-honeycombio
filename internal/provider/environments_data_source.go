package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	v2client "github.com/honeycombio/terraform-provider-honeycombio/client/v2"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/hashcode"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &environmentsDataSource{}
	_ datasource.DataSourceWithConfigure = &environmentsDataSource{}
)

func NewEnvironmentsDataSource() datasource.DataSource {
	return &environmentsDataSource{}
}

type environmentsDataSource struct {
	client *v2client.Client
}

func (d *environmentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environments"
}

func (d *environmentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	w := getClientFromDatasourceRequest(&req)
	if w == nil {
		return
	}

	c, err := w.V2Client()
	if err != nil || c == nil {
		resp.Diagnostics.AddError("Failed to configure client", err.Error())
		return
	}
	d.client = c
}

func (d *environmentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the Environments in a Team.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				Optional: false,
				Required: false,
			},
			"ids": schema.ListAttribute{
				Description: "The list returned of Environments IDs.",
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

func (d *environmentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.EnvironmentsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pager, err := d.client.Environments.List(ctx)
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Listing Environments", err) {
		return
	}
	envs := []*v2client.Environment{}
	for pager.HasNext() {
		items, err := pager.Next(ctx)
		if helper.AddDiagnosticOnError(&resp.Diagnostics, "Listing Environments", err) {
			return
		}
		envs = append(envs, items...)
	}

	// Create a filter group with all filters (implicit AND logic)
	filterGroup, err := models.NewFilterGroup(data.DetailFilter)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Environment filter group", err.Error())
		return
	}

	for _, e := range envs {
		if filterGroup == nil || filterGroup.Match(e) {
			data.IDs = append(data.IDs, types.StringValue(e.ID))
		}
	}
	data.ID = types.StringValue(hashcode.StringValues(data.IDs))

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
