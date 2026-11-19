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
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/hashcode"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &usersDataSource{}
	_ datasource.DataSourceWithConfigure = &usersDataSource{}
)

func NewUsersDataSource() datasource.DataSource {
	return &usersDataSource{}
}

type usersDataSource struct {
	client *v2client.Client
}

func (d *usersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *usersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the users on the Team, or the users in one Group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"group_id": schema.StringAttribute{
				Description: "The ID of a Group. If set, only the users in that Group are returned.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"users": schema.ListNestedAttribute{
				Description: "The users, in order of their ID.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the user.",
							Computed:    true,
						},
						"email": schema.StringAttribute{
							Description: "The user's email address, in lower case.",
							Computed:    true,
						},
						"role": schema.StringAttribute{
							Description: "The user's role on the Team: `owner`, `member` or `read-only`.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *usersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *usersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.UsersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := []v2client.ListOption{v2client.PageSize(100)}
	if !data.GroupID.IsNull() {
		opts = append(opts, v2client.Filter("group", data.GroupID.ValueString()))
	}
	pager, err := d.client.Users.List(ctx, opts...)
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Listing Honeycomb Users", err) {
		return
	}

	data.Users = []models.UserModel{}
	ids := []types.String{}
	for pager.HasNext() {
		users, err := pager.Next(ctx)
		var detailedErr hny.DetailedError
		if errors.As(err, &detailedErr) && detailedErr.IsNotFound() {
			if data.GroupID.IsNull() {
				resp.Diagnostics.AddError("Error Listing Honeycomb Users", "This Team can't list users: the API returned 404 Not Found.")
			} else {
				resp.Diagnostics.AddAttributeError(
					path.Root("group_id"),
					"Error Listing Honeycomb Users",
					fmt.Sprintf("No Group has the ID %q, or this Team can't list users.", data.GroupID.ValueString()),
				)
			}
			return
		}
		if helper.AddDiagnosticOnError(&resp.Diagnostics, "Listing Honeycomb Users", err) {
			return
		}
		for _, u := range users {
			data.Users = append(data.Users, models.UserModel{
				ID:    types.StringValue(u.ID),
				Email: types.StringValue(u.Email),
				Role:  types.StringValue(u.Role),
			})
			ids = append(ids, types.StringValue(u.ID))
		}
	}
	data.ID = types.StringValue(hashcode.StringValues(ids))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
