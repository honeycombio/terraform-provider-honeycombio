package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	v2client "github.com/honeycombio/terraform-provider-honeycombio/client/v2"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAcc_EnvironmentDataSource(t *testing.T) {
	ctx := context.Background()
	c := testAccV2Client(t)

	env, err := c.Environments.Create(ctx, &v2client.Environment{
		Name:        test.RandomStringWithPrefix("test.", 20),
		Description: helper.ToPtr("test environment"),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		c.Environments.Update(ctx, &v2client.Environment{
			ID: env.ID,
			Settings: &v2client.EnvironmentSettings{
				DeleteProtected: helper.ToPtr(false),
			},
		})
		c.Environments.Delete(ctx, env.ID)
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "honeycombio_environment" "test" {
  id = "%s"
}`, env.ID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_environment.test", "id", env.ID),
					resource.TestCheckResourceAttr("data.honeycombio_environment.test", "name", env.Name),
					resource.TestCheckResourceAttr("data.honeycombio_environment.test", "slug", env.Slug),
					resource.TestCheckResourceAttr("data.honeycombio_environment.test", "color", *env.Color),
					resource.TestCheckResourceAttr("data.honeycombio_environment.test", "description", *env.Description),
					resource.TestCheckResourceAttr("data.honeycombio_environment.test", "delete_protected", "true"),
				),
			},
		},
	})

}
