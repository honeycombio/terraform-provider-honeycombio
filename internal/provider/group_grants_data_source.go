package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	hny "github.com/honeycombio/terraform-provider-honeycombio/client"
	v2client "github.com/honeycombio/terraform-provider-honeycombio/client/v2"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &groupGrantsDataSource{}
	_ datasource.DataSourceWithConfigure = &groupGrantsDataSource{}
)

func NewGroupGrantsDataSource() datasource.DataSource {
	return &groupGrantsDataSource{}
}

type groupGrantsDataSource struct {
	client *v2client.Client
}

func (*groupGrantsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_grants"
}

func (*groupGrantsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the Datasets and the Environments that a Honeycomb Group is granted.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the Group.",
				Computed:    true,
			},
			"group_id": schema.StringAttribute{
				Description: "The ID of the Group.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"dataset_ids": schema.SetAttribute{
				Description: "The unique IDs of every Dataset that the Group is granted, including grants made outside Terraform.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"environment_ids": schema.SetAttribute{
				Description: "The IDs of every Environment that the Group is granted, including grants made outside Terraform.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *groupGrantsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *groupGrantsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.GroupGrantsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := listGroupGrants(ctx, d.client, data.GroupID.ValueString())
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.IsNotFound() {
		resp.Diagnostics.AddAttributeError(
			path.Root("group_id"),
			"Error Reading Honeycomb Group Grants",
			fmt.Sprintf("The API found no Group %q.", data.GroupID.ValueString()),
		)
		return
	}
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Reading Honeycomb Group Grants", err) {
		return
	}

	data.ID = data.GroupID
	resp.Diagnostics.Append(current.storeIn(ctx, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
