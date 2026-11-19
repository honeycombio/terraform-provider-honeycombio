package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	hny "github.com/honeycombio/terraform-provider-honeycombio/client"
	v2client "github.com/honeycombio/terraform-provider-honeycombio/client/v2"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/validation"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &userDataSource{}
	_ datasource.DataSourceWithConfigure = &userDataSource{}
)

func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

type userDataSource struct {
	client *v2client.Client
}

func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a user on the Team by their email address.",
		Attributes: map[string]schema.Attribute{
			"email": schema.StringAttribute{
				Description: "The user's email address. It matches without regard to case.",
				Required:    true,
				Validators: []validator.String{
					validation.IsEmailAddress(),
				},
			},
			"id": schema.StringAttribute{
				Description: "The ID of the user.",
				Computed:    true,
			},
			"role": schema.StringAttribute{
				Description: "The user's role on the Team: `owner`, `member` or `read-only`.",
				Computed:    true,
			},
		},
	}
}

func (d *userDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.UserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// an email matches at most one user, so the first page is the whole answer
	pager, err := d.client.Users.List(ctx, v2client.Filter("email", data.Email.ValueString()))
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Looking up Honeycomb User", err) {
		return
	}
	users, err := pager.Next(ctx)
	var detailedErr hny.DetailedError
	if errors.As(err, &detailedErr) && detailedErr.IsNotFound() {
		resp.Diagnostics.AddError("Error Looking up Honeycomb User", "This Team can't list users: the API returned 404 Not Found.")
		return
	}
	if helper.AddDiagnosticOnError(&resp.Diagnostics, "Looking up Honeycomb User", err) {
		return
	}
	if len(users) == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("email"),
			"No User found",
			fmt.Sprintf("No user on the Team has the email %q. Pending invites are not listed.", data.Email.ValueString()),
		)
		return
	}

	data.ID = types.StringValue(users[0].ID)
	data.Role = types.StringValue(users[0].Role)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
