package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/stretchr/testify/require"
)

func TestAcc_AuthMetadataData(t *testing.T) {
	ctx := t.Context()
	c := testAccClient(t)

	metadata, err := c.Auth.List(ctx)
	require.NoError(t, err)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
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
