package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &datasetDataSource{}
	_ datasource.DataSourceWithConfigure = &datasetDataSource{}
)

func NewDatasetDataSource() datasource.DataSource {
	return &datasetDataSource{}
}

type datasetDataSource struct {
	client *client.Client
}

func (d *datasetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataset"
}

func (d *datasetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the details of a single Dataset.",
		Attributes: map[string]schema.Attribute{
			"slug": schema.StringAttribute{
				Description: "The Slug of the Dataset to fetch.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of the Dataset.",
				Computed:    true,
				Optional:    false,
				Required:    false,
			},
			"name": schema.StringAttribute{
				Description: "The name of the Dataset.",
				Computed:    true,
				Optional:    false,
				Required:    false,
			},
			"description": schema.StringAttribute{
				Description: "The Dataset's description.",
				Computed:    true,
				Optional:    false,
				Required:    false,
			},
			"expand_json_depth": schema.Int32Attribute{
				Description: "The Dataset's maximum unpacking depth of nested JSON fields.",
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
			"created_at": schema.StringAttribute{
				Description: "ISO8601-formatted time the dataset was created.",
				Computed:    true,
				Required:    false,
				Optional:    false,
			},
			"last_written_at": schema.StringAttribute{
				Description: "ISO8601-formatted time the dataset was last written to (received event data).",
				Computed:    true,
				Required:    false,
				Optional:    false,
			},
		},
	}
}

func (d *datasetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *datasetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.DatasetResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ds, err := d.client.Datasets.Get(ctx, data.Slug.ValueString())
	if helper.AddDiagnosticOnError(&resp.Diagnostics,
		fmt.Sprintf("Looking up Dataset %q", data.Slug.ValueString()), err) {
		return
	}

	data.ID = types.StringValue(ds.Slug)
	data.Slug = types.StringValue(ds.Slug)
	data.Name = types.StringValue(ds.Name)
	data.Description = types.StringValue(ds.Description)
	data.ExpandJSONDepth = types.Int32Value(int32(ds.ExpandJSONDepth))
	data.DeleteProtected = types.BoolPointerValue(ds.Settings.DeleteProtected)
	data.CreatedAt = types.StringValue(ds.CreatedAt.Format(time.RFC3339))
	data.LastWrittenAt = types.StringValue(ds.LastWrittenAt.Format(time.RFC3339))

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
