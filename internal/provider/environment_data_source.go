package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	v2client "github.com/honeycombio/terraform-provider-honeycombio/client/v2"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/filter"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &environmentDataSource{}
	_ datasource.DataSourceWithConfigure = &environmentDataSource{}
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
		Description: `Fetches the details of a single Environment.

Note: Terraform will fail unless exactly one environment is returned by the search.
Ensure that your search is specific enough to return a single environment only.
If you want to match multiple environments, use the 'honeycombio_environments' data source instead.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the Environment to fetch.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("detail_filter")),
				},
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
		Blocks: map[string]schema.Block{
			"detail_filter": detailFilterSchema(),
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
	var data models.EnvironmentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var env *v2client.Environment
	if !data.ID.IsNull() {
		env, err = d.client.Environments.Get(ctx, data.ID.ValueString())
		if helper.AddDiagnosticOnError(&resp.Diagnostics,
			fmt.Sprintf("Looking up Environment %q", data.ID.ValueString()), err) {
			return
		}
	} else {
		// we're using the detail filter to find the environment
		var envFilter *filter.DetailFilter
		if len(data.DetailFilter) > 0 {
			envFilter, err = data.DetailFilter[0].NewFilter()
			if err != nil {
				resp.Diagnostics.AddError("Unable to create Environment filter", err.Error())
				return
			}
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

		matched := make([]*v2client.Environment, 0, len(envs))
		for _, e := range envs {
			if !envFilter.MatchName(e.Name) {
				continue
			}
			matched = append(matched, e)
		}

		if len(matched) == 0 {
			resp.Diagnostics.AddError(
				"No Environments found",
				"Your filter returned no matches.",
			)
			return
		}
		if len(matched) > 1 {
			resp.Diagnostics.AddError(
				"Multiple Environments found",
				"Please filter by ID or use a more specific detail filter.",
			)
			return
		}
		env = matched[0]
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
