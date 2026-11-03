package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	v2client "github.com/honeycombio/terraform-provider-honeycombio/client/v2"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &authMetadataDataSource{}
	_ datasource.DataSourceWithConfigure = &authMetadataDataSource{}
)

func NewAuthMetadataDataSource() datasource.DataSource {
	return &authMetadataDataSource{}
}

// authMetadataDataSource is the data source implementation.
type authMetadataDataSource struct {
	// client is used when the provider is configured with an API Key.
	client *client.Client
	// v2client is used when it is configured only with a Key ID and Secret pair,
	// as the v1 API does not accept Management Keys.
	v2client *v2client.Client
}

type authMetadataDataSourceModel struct {
	APIKeyAccess struct {
		Boards     types.Bool `tfsdk:"boards"`
		Columns    types.Bool `tfsdk:"columns"`
		Datasets   types.Bool `tfsdk:"datasets"`
		Events     types.Bool `tfsdk:"events"`
		Markers    types.Bool `tfsdk:"markers"`
		Queries    types.Bool `tfsdk:"queries"`
		Recipients types.Bool `tfsdk:"recipients"`
		SLOs       types.Bool `tfsdk:"slos"`
		Triggers   types.Bool `tfsdk:"triggers"`
	} `tfsdk:"api_key_access"`
	Environment struct {
		ID      types.String `tfsdk:"id"`
		Classic types.Bool   `tfsdk:"classic"`
		Name    types.String `tfsdk:"name"`
		Slug    types.String `tfsdk:"slug"`
	} `tfsdk:"environment"`
	Team struct {
		Name types.String `tfsdk:"name"`
		Slug types.String `tfsdk:"slug"`
	} `tfsdk:"team"`
}

func (d *authMetadataDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auth_metadata"
}

func (d *authMetadataDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retreives information about the API key used to authenticate the provider.\n\n" +
			"If the provider is configured with an API Key (`api_key`), the information comes from the v1 API. " +
			"If it is configured only with a Key ID and Secret pair, it comes from the v2 API and `api_key_access` is not set.",
		Blocks: map[string]schema.Block{
			"api_key_access": schema.SingleNestedBlock{
				Description: "The authorizations granted for the API key used to authenticate the provider. Only set if the provider is configured with an API Key. See https://docs.honeycomb.io/working-with-your-data/settings/api-keys/ for more information.",
				Attributes: map[string]schema.Attribute{
					"boards": schema.BoolAttribute{
						Description: "Whether this API key can create and manage Boards.",
						Computed:    true,
						Optional:    false,
						Required:    false,
					},
					"columns": schema.BoolAttribute{
						Description: "Whether this API key can create and manage Queries, Columns, Derived Columns, and Query Annotations.",
						Computed:    true,
						Optional:    false,
						Required:    false,
					},
					"datasets": schema.BoolAttribute{
						Description: "Whether this API key can create and manage Datasets.",
						Computed:    true,
						Optional:    false,
						Required:    false,
					},
					"events": schema.BoolAttribute{
						Description: "Whether this API key can send events to Honeycomb.",
						Computed:    true,
						Optional:    false,
						Required:    false,
					},
					"markers": schema.BoolAttribute{
						Description: "Whether this API key can create and manage Markers.",
						Computed:    true,
						Optional:    false,
						Required:    false,
					},
					"queries": schema.BoolAttribute{
						Description: "Whether this API key can execute existing Queries via the Query Data API.",
						Computed:    true,
						Optional:    false,
						Required:    false,
					},
					"recipients": schema.BoolAttribute{
						Description: "Whether this API key can create and manage Recipients.",
						Computed:    true,
						Optional:    false,
						Required:    false,
					},
					"slos": schema.BoolAttribute{
						Description: "Whether this API key can create and manage SLOs and Burn Alerts.",
						Computed:    true,
						Optional:    false,
						Required:    false,
					},
					"triggers": schema.BoolAttribute{
						Description: "Whether this API key can create and manage Triggers.",
						Computed:    true,
						Optional:    false,
						Required:    false,
					},
				},
			},
			"environment": schema.SingleNestedBlock{
				Description: "Information about the Environment the API key is scoped to. " +
					"If the provider is configured with a team-scoped Management Key, the Environment is not set.",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "The ID of the Environment. Only set if the provider is configured with a Key ID and Secret pair.",
						Computed:    true,
						Optional:    false,
						Required:    false,
					},
					"classic": schema.BoolAttribute{
						Description: "Whether the Environment is a Classic Environment.",
						Computed:    true,
						Optional:    false,
						Required:    false,
					},
					"name": schema.StringAttribute{
						Description: "The name of the Environment. For Classic environments, this will be null.",
						Computed:    true,
						Optional:    false,
						Required:    false,
					},
					"slug": schema.StringAttribute{
						Description: "The slug of the Environment. For Classic environments, this will be null.",
						Computed:    true,
						Optional:    false,
						Required:    false,
					},
				},
			},
			"team": schema.SingleNestedBlock{
				Description: "Information about the Team the API key is scoped to.",
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Description: "The name of the Team.",
						Computed:    true,
						Optional:    false,
						Required:    false,
					},
					"slug": schema.StringAttribute{
						Description: "The slug of the Team.",
						Computed:    true,
						Optional:    false,
						Required:    false,
					},
				},
			},
		},
	}
}

func (d *authMetadataDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	w := getClientFromDatasourceRequest(&req)
	if w == nil {
		return
	}

	// The provider guarantees at least one of the clients is configured.
	d.client = w.v1client
	d.v2client = w.v2client
}

func (d *authMetadataDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data authMetadataDataSourceModel

	if d.client == nil {
		d.readV2(ctx, &data, resp)
		return
	}

	metadata, err := d.client.Auth.List(ctx)
	if helper.AddDiagnosticOnError(&resp.Diagnostics,
		"Listing Auth Metadata",
		err) {
		return
	}

	data.APIKeyAccess.Boards = types.BoolValue(metadata.APIKeyAccess.Boards)
	data.APIKeyAccess.Columns = types.BoolValue(metadata.APIKeyAccess.Columns)
	data.APIKeyAccess.Datasets = types.BoolValue(metadata.APIKeyAccess.CreateDatasets)
	data.APIKeyAccess.Events = types.BoolValue(metadata.APIKeyAccess.Events)
	data.APIKeyAccess.Markers = types.BoolValue(metadata.APIKeyAccess.Markers)
	data.APIKeyAccess.Queries = types.BoolValue(metadata.APIKeyAccess.Queries)
	data.APIKeyAccess.Recipients = types.BoolValue(metadata.APIKeyAccess.Recipients)
	data.APIKeyAccess.SLOs = types.BoolValue(metadata.APIKeyAccess.SLOs)
	data.APIKeyAccess.Triggers = types.BoolValue(metadata.APIKeyAccess.Triggers)

	data.Team.Name = types.StringValue(metadata.Team.Name)
	data.Team.Slug = types.StringValue(metadata.Team.Slug)

	// Classic environments don't have a name or slug
	if metadata.Environment.Slug == "" {
		data.Environment.Classic = types.BoolValue(true)
		data.Environment.Name = types.StringNull()
		data.Environment.Slug = types.StringNull()
	} else {
		data.Environment.Classic = types.BoolValue(false)
		data.Environment.Name = types.StringValue(metadata.Environment.Name)
		data.Environment.Slug = types.StringValue(metadata.Environment.Slug)
	}

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// readV2 fills the data from the v2 API, which unlike v1 accepts Management Keys.
func (d *authMetadataDataSource) readV2(ctx context.Context, data *authMetadataDataSourceModel, resp *datasource.ReadResponse) {
	metadata, err := d.v2client.AuthInfo(ctx)
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Retrieving Auth Metadata", err) {
		return
	}

	data.Team.Name = types.StringValue(metadata.Team.Name)
	data.Team.Slug = types.StringValue(metadata.Team.Slug)

	// only environment-scoped keys have an Environment.
	// All of the other attributes are left null: Management Keys have scopes,
	// not the API Key access of v1 keys, and a team-scoped key isn't Classic.
	if metadata.Environment != nil {
		data.Environment.ID = types.StringValue(metadata.Environment.ID)
		data.Environment.Classic = types.BoolValue(false)
		data.Environment.Name = types.StringValue(metadata.Environment.Name)
		data.Environment.Slug = types.StringValue(metadata.Environment.Slug)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}
