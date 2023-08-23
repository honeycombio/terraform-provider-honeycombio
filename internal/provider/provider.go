package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/log"
)

// Ensure HoneycombioProvider satisfies various provider interfaces.
var _ provider.Provider = &HoneycombioProvider{}

type HoneycombioProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// HoneycombioProviderModel describes the provider data model.
type HoneycombioProviderModel struct {
	APIKey types.String `tfsdk:"api_key"`
	APIUrl types.String `tfsdk:"api_url"`
	Debug  types.Bool   `tfsdk:"debug"`
}

func New(version string) provider.Provider {
	return &HoneycombioProvider{
		version: version,
	}
}

func (p *HoneycombioProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The Honeycomb API key to use. It can also be set using HONEYCOMB_API_KEY or HONEYCOMBIO_APIKEY environment variables.",
				Optional:            true,
				Sensitive:           true,
			},
			"api_url": schema.StringAttribute{
				MarkdownDescription: "Override the URL of the Honeycomb.io API. Defaults to https://api.honeycomb.io.",
				Optional:            true,
			},
			"debug": schema.BoolAttribute{
				MarkdownDescription: "Enable the API client's debug logs. By default, a `TF_LOG` setting of debug or higher will enable this.",
				Optional:            true,
			},
		},
	}
}

func (p *HoneycombioProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewBurnAlertResource,
		NewTriggerResource,
	}
}

func (p *HoneycombioProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDerivedColumnDataSource,
		NewDerivedColumnsDataSource,
		NewSLODataSource,
		NewSLOsDataSource,
	}
}

func (p *HoneycombioProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "honeycombio"
	resp.Version = p.version
}

func (p *HoneycombioProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config HoneycombioProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown Honeycomb API Key",
			"The provider cannot create the Honeycomb client as there is an unknown configuration value for the Honeycomb API Key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HONEYCOMB_API_KEY environment variable.",
		)
	}

	apiKey := os.Getenv("HONEYCOMB_API_KEY")
	if apiKey == "" {
		// fall through to legacy env var
		apiKey = os.Getenv("HONEYCOMBIO_APIKEY")
	}
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}
	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Honeycomb API Key",
			"The provider cannot create the Honeycomb API client as there is a missing or empty value for the Honeycomb API Key. "+
				"Set the value in the configuration or use the HONEYCOMB_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	debug := log.IsDebugOrHigher()
	if !config.Debug.IsNull() {
		debug = config.Debug.ValueBool()
	}

	client, err := client.NewClient(&client.Config{
		APIKey:    apiKey,
		APIUrl:    config.APIUrl.ValueString(),
		Debug:     debug,
		UserAgent: fmt.Sprintf("Terraform/%s terraform-provider-honeycombio/%s", req.TerraformVersion, p.version),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Honeycomb API Client",
			"An unexpected error occurred when creating the Honeycomb API client.\n\n "+
				"Honeycomb Client Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func getClientFromDatasourceRequest(req *datasource.ConfigureRequest) *client.Client {
	if req.ProviderData == nil {
		return nil
	}
	return req.ProviderData.(*client.Client)
}

func getClientFromResourceRequest(req *resource.ConfigureRequest) *client.Client {
	if req.ProviderData == nil {
		return nil
	}
	return req.ProviderData.(*client.Client)
}
