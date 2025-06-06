package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/hashcode"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &datasetsDataSource{}
	_ datasource.DataSourceWithConfigure = &datasetsDataSource{}
)

func NewDatasetsDataSource() datasource.DataSource {
	return &datasetsDataSource{}
}

type datasetsDataSource struct {
	client *client.Client
}

func (d *datasetsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datasets"
}

func (d *datasetsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *datasetsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the Datasets in an Environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				Optional: false,
				Required: false,
			},
			"starts_with": schema.StringAttribute{
				DeprecationMessage: "Use the `detail_filter` block instead.",
				Description:        "The prefix to filter the Dataset Names by.",
				Optional:           true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("detail_filter")),
				},
			},
			"names": schema.ListAttribute{
				Description: "The list returned of Dataset Names.",
				Computed:    true,
				Optional:    false,
				Required:    false,
				ElementType: types.StringType,
			},
			"slugs": schema.ListAttribute{
				Description: "The list returned of Dataset Slugs.",
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

func (d *datasetsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.DatasetsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	datasets, err := d.client.Datasets.List(ctx)
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Listing Datasets", err) {
		return
	}

	filterGroup, err := models.NewFilterGroup(data.DetailFilter)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Dataset filter group", err.Error())
		return
	}

	for _, e := range datasets {
		datasetResource := datasetToResourceModel(e)

		if filterGroup == nil || filterGroup.Match(datasetResource) {
			data.Names = append(data.Names, types.StringValue(e.Name))
			data.Slugs = append(data.Slugs, types.StringValue(e.Slug))
		}
	}
	data.ID = types.StringValue(hashcode.StringValues(data.Slugs))

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// datasetToResourceModel converts a client.Dataset to a DatasetResourceModel
func datasetToResourceModel(ds client.Dataset) *models.DatasetResourceModel {
	return &models.DatasetResourceModel{
		ID:              types.StringValue(ds.Slug),
		Slug:            types.StringValue(ds.Slug),
		Name:            types.StringValue(ds.Name),
		Description:     types.StringValue(ds.Description),
		ExpandJSONDepth: types.Int32Value(int32(ds.ExpandJSONDepth)),
		DeleteProtected: types.BoolPointerValue(ds.Settings.DeleteProtected),
		CreatedAt:       types.StringValue(ds.CreatedAt.Format(time.RFC3339)),
		LastWrittenAt:   types.StringValue(ds.LastWrittenAt.Format(time.RFC3339)),
	}
}
