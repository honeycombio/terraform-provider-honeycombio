package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	v2client "github.com/honeycombio/terraform-provider-honeycombio/client/v2"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &groupDataSource{}
	_ datasource.DataSourceWithConfigure = &groupDataSource{}
)

func NewGroupDataSource() datasource.DataSource {
	return &groupDataSource{}
}

type groupDataSource struct {
	client *v2client.Client
}

func (d *groupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (d *groupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a Group on the Team by its name.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the Group. It must match exactly, including case.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.UTF8LengthBetween(1, 255),
				},
			},
			"id": schema.StringAttribute{
				Description: "The ID of the Group.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The Group's description.",
				Computed:    true,
			},
		},
	}
}

func (d *groupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *groupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.GroupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// the API cannot filter Groups by name, so read the pages until a name matches.
	// A name is unique within the Team, so the first match is the only one.
	pager, err := d.client.Groups.List(ctx)
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Looking up Honeycomb Group", err) {
		return
	}
	for pager.HasNext() {
		groups, err := pager.Next(ctx)
		if helper.AddDiagnosticOnError(&resp.Diagnostics, "Looking up Honeycomb Group", err) {
			return
		}
		for _, g := range groups {
			if g.Name == data.Name.ValueString() {
				data = flattenGroup(g)
				resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
				return
			}
		}
	}

	resp.Diagnostics.AddAttributeError(
		path.Root("name"),
		"No Group found",
		fmt.Sprintf("No Group on the Team has the name %q. The name must match exactly, including case.", data.Name.ValueString()),
	)
}
