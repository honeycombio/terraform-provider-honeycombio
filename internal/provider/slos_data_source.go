package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/filter"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/hashcode"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/validation"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &slosDataSource{}
	_ datasource.DataSourceWithConfigure = &slosDataSource{}
)

func NewSLOsDataSource() datasource.DataSource {
	return &slosDataSource{}
}

// slosDataSource is the data source implementation.
type slosDataSource struct {
	client *client.Client
}

type slosDataSourceModel struct {
	ID           types.String       `tfsdk:"id"`
	Dataset      types.String       `tfsdk:"dataset"`
	DetailFilter []slosDetailFilter `tfsdk:"detail_filter"`
	SLOs         []types.String     `tfsdk:"ids"`
}

type slosDetailFilter struct {
	Name       types.String `tfsdk:"name"`
	Value      types.String `tfsdk:"value"`
	ValueRegex types.String `tfsdk:"value_regex"`
}

func (f *slosDetailFilter) SLOFilter() (*filter.SLODetailFilter, error) {
	if f == nil {
		return nil, nil
	}
	return filter.NewDetailSLOFilter(f.Name.ValueString(), f.Value.ValueString(), f.ValueRegex.ValueString())
}

func (d *slosDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_slos"
}

func (d *slosDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the SLOs in a dataset.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"dataset": schema.StringAttribute{
				Description: "The dataset to fetch the SLOs from.",
				Required:    true,
			},
			"ids": schema.ListAttribute{
				Description: "The list of SLO IDs.",
				Computed:    true,
				Optional:    false,
				Required:    false,
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"detail_filter": schema.ListNestedBlock{
				Description: "Attributes to filter the SLOs with. `name` must be set when providing a filter.",
				Validators:  []validator.List{listvalidator.SizeAtMost(1)},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The name of the detail field to filter by.",
							Validators:  []validator.String{stringvalidator.OneOf("name")},
						},
						"value": schema.StringAttribute{
							Optional:    true,
							Description: "The value of the detail field to match on.",
							Validators: []validator.String{
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("value_regex")),
							},
						},
						"value_regex": schema.StringAttribute{
							Optional:    true,
							Description: "A regular expression string to apply to the value of the detail field to match on.",
							Validators: []validator.String{
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("value")),
								validation.IsValidRegExp(),
							},
						},
					},
				},
			},
		},
	}
}

func (d *slosDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	d.client = getClientFromDatasourceRequest(&req)
}

func (d *slosDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data slosDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	slos, err := d.client.SLOs.List(ctx, data.Dataset.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to list SLOs", err.Error())
		return
	}

	var sloFilter *filter.SLODetailFilter
	if len(data.DetailFilter) > 0 {
		sloFilter, err = data.DetailFilter[0].SLOFilter()
		if err != nil {
			resp.Diagnostics.AddError("Unable to create SLO filter", err.Error())
			return
		}
	}
	for _, s := range slos {
		if sloFilter != nil && !sloFilter.Match(s) {
			continue
		}
		data.SLOs = append(data.SLOs, types.StringValue(s.ID))
	}
	data.ID = types.StringValue(hashcode.StringValues(data.SLOs))

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
