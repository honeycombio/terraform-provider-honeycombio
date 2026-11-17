package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/stretchr/testify/require"

	v2client "github.com/honeycombio/terraform-provider-honeycombio/client/v2"
)

func TestAcc_AuthMetadataData(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	c := testAccClient(t)

	metadata, err := c.Auth.List(ctx)
	require.NoError(t, err)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: `data "honeycombio_auth_metadata" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_auth_metadata.test", "team.name", metadata.Team.Name),
					resource.TestCheckResourceAttr("data.honeycombio_auth_metadata.test", "team.slug", metadata.Team.Slug),
					resource.TestCheckResourceAttr("data.honeycombio_auth_metadata.test", "environment.classic", fmt.Sprintf("%v", c.IsClassic(ctx))),
					resource.TestCheckResourceAttr("data.honeycombio_auth_metadata.test", "api_key_access.boards", fmt.Sprintf("%v", metadata.APIKeyAccess.Boards)),
					resource.TestCheckResourceAttr("data.honeycombio_auth_metadata.test", "api_key_access.columns", fmt.Sprintf("%v", metadata.APIKeyAccess.Columns)),
					resource.TestCheckResourceAttr("data.honeycombio_auth_metadata.test", "api_key_access.datasets", fmt.Sprintf("%v", metadata.APIKeyAccess.CreateDatasets)),
					resource.TestCheckResourceAttr("data.honeycombio_auth_metadata.test", "api_key_access.events", fmt.Sprintf("%v", metadata.APIKeyAccess.Events)),
					resource.TestCheckResourceAttr("data.honeycombio_auth_metadata.test", "api_key_access.markers", fmt.Sprintf("%v", metadata.APIKeyAccess.Markers)),
					resource.TestCheckResourceAttr("data.honeycombio_auth_metadata.test", "api_key_access.queries", fmt.Sprintf("%v", metadata.APIKeyAccess.Queries)),
					resource.TestCheckResourceAttr("data.honeycombio_auth_metadata.test", "api_key_access.recipients", fmt.Sprintf("%v", metadata.APIKeyAccess.Recipients)),
					resource.TestCheckResourceAttr("data.honeycombio_auth_metadata.test", "api_key_access.slos", fmt.Sprintf("%v", metadata.APIKeyAccess.SLOs)),
					resource.TestCheckResourceAttr("data.honeycombio_auth_metadata.test", "api_key_access.triggers", fmt.Sprintf("%v", metadata.APIKeyAccess.Triggers)),
				),
			},
		},
	})
}

// With only a Key ID and Secret pair the data comes from the v2 API.
func TestAcc_AuthMetadataData_V2(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("team-scoped Management Key", func(t *testing.T) {
		metadata, err := testAccV2Client(t).AuthInfo(ctx)
		require.NoError(t, err)

		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheckV2API(t),
			ProtoV6ProviderFactories: testAccProtoV6MuxServerFactory,
			Steps: []resource.TestStep{
				{
					Config: testAccProviderV2Only(os.Getenv("HONEYCOMB_KEY_ID"), os.Getenv("HONEYCOMB_KEY_SECRET")) +
						`data "honeycombio_auth_metadata" "test" {}`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.honeycombio_auth_metadata.test", "team.name", metadata.Team.Name),
						resource.TestCheckResourceAttr("data.honeycombio_auth_metadata.test", "team.slug", metadata.Team.Slug),
						resource.TestCheckNoResourceAttr("data.honeycombio_auth_metadata.test", "environment.id"),
						resource.TestCheckNoResourceAttr("data.honeycombio_auth_metadata.test", "environment.slug"),
						resource.TestCheckNoResourceAttr("data.honeycombio_auth_metadata.test", "api_key_access.boards"),
					),
				},
			},
		})
	})

	t.Run("environment-scoped Management Key", func(t *testing.T) {
		id, secret := testAccEnvScopedKey(t)
		c, err := v2client.NewClientWithConfig(&v2client.Config{APIKeyID: id, APIKeySecret: secret})
		require.NoError(t, err)
		metadata, err := c.AuthInfo(ctx)
		require.NoError(t, err)
		require.NotNil(t, metadata.Environment, "HONEYCOMB_ENV_KEY_ID is not an environment-scoped key")

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6MuxServerFactory,
			Steps: []resource.TestStep{
				{
					Config: testAccProviderV2Only(id, secret) + `data "honeycombio_auth_metadata" "test" {}`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.honeycombio_auth_metadata.test", "team.slug", metadata.Team.Slug),
						resource.TestCheckResourceAttr("data.honeycombio_auth_metadata.test", "environment.id", metadata.Environment.ID),
						resource.TestCheckResourceAttr("data.honeycombio_auth_metadata.test", "environment.name", metadata.Environment.Name),
						resource.TestCheckResourceAttr("data.honeycombio_auth_metadata.test", "environment.slug", metadata.Environment.Slug),
						resource.TestCheckResourceAttr("data.honeycombio_auth_metadata.test", "environment.classic", "false"),
					),
				},
			},
		})
	})
}
