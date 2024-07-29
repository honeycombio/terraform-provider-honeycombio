package provider

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	v2client "github.com/honeycombio/terraform-provider-honeycombio/client/v2"
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
	APIKey    types.String `tfsdk:"api_key"`
	KeyID     types.String `tfsdk:"api_key_id"`
	KeySecret types.String `tfsdk:"api_key_secret"`
	APIUrl    types.String `tfsdk:"api_url"`
	Debug     types.Bool   `tfsdk:"debug"`
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
				MarkdownDescription: "The Honeycomb API key to use. It can also be set via the `HONEYCOMB_API_KEY` or `HONEYCOMBIO_APIKEY` environment variables.",
				Optional:            true,
				Sensitive:           true,
			},
			"api_key_id": schema.StringAttribute{
				MarkdownDescription: "The ID portion of the Honeycomb Management API key to use. It can also be set via the `HONEYCOMB_KEY_ID` environment variable.",
				Optional:            true,
				Sensitive:           false,
			},
			"api_key_secret": schema.StringAttribute{
				MarkdownDescription: "The secret portion of the Honeycomb Management API key to use. It can also be set via the `HONEYCOMB_KEY_SECRET` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"api_url": schema.StringAttribute{
				MarkdownDescription: "Override the URL of the Honeycomb API. Defaults to `https://api.honeycomb.io`. It can also be set via the `HONEYCOMB_API_ENDPOINT` environment variable.",
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
		NewQueryResource,
		NewAPIKeyResource,
		NewEnvironmentResource,
	}
}

func (p *HoneycombioProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAuthMetadataDataSource,
		NewDerivedColumnDataSource,
		NewDerivedColumnsDataSource,
		NewEnvironmentDataSource,
		NewEnvironmentsDataSource,
		NewSLODataSource,
		NewSLOsDataSource,
		NewQuerySpecDataSource,
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
	if config.KeyID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key_id"),
			"Unknown Honeycomb API Key ID",
			"The provider cannot create the Honeycomb client as there is an unknown configuration value for the Honeycomb API Key ID. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HONEYCOMB_KEY_ID environment variable.",
		)
	}
	if config.KeySecret.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key_secret"),
			"Unknown Honeycomb API Key Secret",
			"The provider cannot create the Honeycomb client as there is an unknown configuration value for the Honeycomb API Key Secret. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HONEYCOMB_KEY_SECRET environment variable.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv(client.DefaultAPIKeyEnv)
	if apiKey == "" {
		// fall through to legacy env var
		//nolint:staticcheck
		apiKey = os.Getenv(client.LegacyAPIKeyEnv)
	}
	keyID := os.Getenv(v2client.DefaultAPIKeyIDEnv)
	keySecret := os.Getenv(v2client.DefaultAPIKeySecretEnv)

	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}
	if !config.KeyID.IsNull() {
		keyID = config.KeyID.ValueString()
	}
	if !config.KeySecret.IsNull() {
		keySecret = config.KeySecret.ValueString()
	}

	initv1Client, initv2Client := false, false
	if apiKey != "" {
		initv1Client = true
	}
	if keyID != "" && keySecret != "" {
		initv2Client = true
	} else if (keyID != "" && keySecret == "") || (keyID == "" && keySecret != "") {
		resp.Diagnostics.AddError(
			"Unable to initialize Honeycomb provider",
			"The provider requires both a Honeycomb API Key ID and Secret. "+
				"Set them both in the configuration or via the HONEYCOMB_KEY_ID and HONEYCOMB_KEY_SECRET"+
				"environment variables. "+
				"If you believe both are already set, ensure the values are not empty.",
		)
		return
	}

	if !initv1Client && !initv2Client {
		resp.Diagnostics.AddError(
			"Unable to initialize Honeycomb provider",
			"The provider requires at least one of a Honeycomb API Key, or the Honeycomb API Key ID and Secret pair. "+
				"Set either HONEYCOMB_API_KEY for v1 APIs, or HONEYCOMB_KEY_ID and HONEYCOMB_KEY_SECRET for v2 APIs. "+
				"If you believe either is already set, ensure the values are not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	debug := log.IsDebugOrHigher()
	if !config.Debug.IsNull() {
		debug = config.Debug.ValueBool()
	}
	userAgent := fmt.Sprintf(
		"Terraform/%s terraform-provider-honeycombio/%s",
		req.TerraformVersion,
		p.version,
	)

	cc := &ConfiguredClient{}
	if initv1Client {
		client, err := client.NewClientWithConfig(&client.Config{
			APIKey:    apiKey,
			APIUrl:    config.APIUrl.ValueString(),
			Debug:     debug,
			UserAgent: userAgent,
		})
		if err != nil {
			resp.Diagnostics.AddError("Unable to create Honeycomb API V1 Client", err.Error())
			return
		}
		cc.v1client = client
	}

	if initv2Client {
		v2client, err := v2client.NewClientWithConfig(&v2client.Config{
			APIKeyID:     keyID,
			APIKeySecret: keySecret,
			BaseURL:      config.APIUrl.ValueString(),
			Debug:        debug,
			UserAgent:    userAgent,
		})
		if err != nil {
			resp.Diagnostics.AddError("Unable to create Honeycomb API V2 Client", err.Error())
			return
		}
		cc.v2client = v2client
	}

	resp.DataSourceData = cc
	resp.ResourceData = cc
}

// ConfiguredClient is a wrapper around the configured Honeycomb API clients.
type ConfiguredClient struct {
	v1client *client.Client
	v2client *v2client.Client
}

func (c *ConfiguredClient) V1Client() (*client.Client, error) {
	if c.v1client == nil {
		return nil, errors.New("No v1 API client configured for this provider. " +
			"Set the `api_key` attribute in the provider's configuration, " +
			"or set the HONEYCOMB_API_KEY environment variable.",
		)
	}
	return c.v1client, nil
}

func (c *ConfiguredClient) V2Client() (*v2client.Client, error) {
	if c.v2client == nil {
		return nil, errors.New("No v2 API client configured for this provider. " +
			"Set the Key ID and Secret pair in the provider's configuration, " +
			"or via the HONEYCOMB_KEY_ID and HONEYCOMB_KEY_SECRET environment variables.",
		)
	}
	return c.v2client, nil
}

func getClientFromDatasourceRequest(req *datasource.ConfigureRequest) *ConfiguredClient {
	if req.ProviderData != nil {
		if c, ok := req.ProviderData.(*ConfiguredClient); ok {
			return c
		}
	}
	// ProviderData hasn't been initialized yet -- so fail gracefully
	return nil
}

func getClientFromResourceRequest(req *resource.ConfigureRequest) *ConfiguredClient {
	if req.ProviderData != nil {
		if c, ok := req.ProviderData.(*ConfiguredClient); ok {
			return c
		}
	}
	// ProviderData hasn't been initialized yet -- so fail gracefully
	return nil
}
