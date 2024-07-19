package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_EnvironmentDataSource(t *testing.T) {
	ctx := context.Background()
	c := testAccV2Client(t)
	env := testAccEnvironment(ctx, t, c)

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
