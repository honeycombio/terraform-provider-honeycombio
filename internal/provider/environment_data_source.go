package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	v2client "github.com/honeycombio/terraform-provider-honeycombio/client/v2"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &sloDataSource{}
	_ datasource.DataSourceWithConfigure = &sloDataSource{}
)

func NewEnvironmentDataSource() datasource.DataSource {
	return &environmentDataSource{}
}

type environmentDataSource struct {
	client *v2client.Client
}

func (d *environmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (d *environmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the details of a single Environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the Environment to fetch.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the Environment.",
				Computed:    true,
				Optional:    false,
				Required:    false,
			},
			"slug": schema.StringAttribute{
				Description: "The slug of the Environment.",
				Computed:    true,
				Optional:    false,
				Required:    false,
			},
			"description": schema.StringAttribute{
				Description: "The Environment's description.",
				Computed:    true,
				Optional:    false,
				Required:    false,
			},
			"color": schema.StringAttribute{
				Description: "The color of the Environment.",
				Computed:    true,
				Optional:    false,
				Required:    false,
			},
			"delete_protected": schema.BoolAttribute{
				Description: "The current delete protection status of the Environment.",
				Computed:    true,
				Optional:    false,
				Required:    false,
			},
		},
	}
}

func (d *environmentDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *environmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.EnvironmentResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env, err := d.client.Environments.Get(ctx, data.ID.ValueString())
	if helper.AddDiagnosticOnError(&resp.Diagnostics,
		fmt.Sprintf("Looking up Environment %q", data.ID.ValueString()), err) {
		return
	}

	data.ID = types.StringValue(env.ID)
	data.Name = types.StringValue(env.Name)
	data.Slug = types.StringValue(env.Slug)
	data.Description = types.StringPointerValue(env.Description)
	data.Color = types.StringPointerValue(env.Color)
	data.DeleteProtected = types.BoolPointerValue(env.Settings.DeleteProtected)

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
