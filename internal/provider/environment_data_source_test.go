package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_EnvironmentDataSource(t *testing.T) {
	ctx := t.Context()
	c := testAccV2Client(t)
	env := testAccEnvironment(ctx, t, c)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "honeycombio_environment" "test_id" {
  id = "%s"
}

data "honeycombio_environment" "test_filter" {
  detail_filter {
    name = "name"
    value = "%s"
  }
}`, env.ID, env.Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_environment.test_id", "id", env.ID),
					resource.TestCheckResourceAttr("data.honeycombio_environment.test_id", "name", env.Name),
					resource.TestCheckResourceAttr("data.honeycombio_environment.test_id", "slug", env.Slug),
					resource.TestCheckResourceAttr("data.honeycombio_environment.test_id", "color", *env.Color),
					resource.TestCheckResourceAttr("data.honeycombio_environment.test_id", "description", *env.Description),
					resource.TestCheckResourceAttr("data.honeycombio_environment.test_id", "delete_protected", "true"),
					resource.TestCheckResourceAttr("data.honeycombio_environment.test_filter", "id", env.ID),
					resource.TestCheckResourceAttr("data.honeycombio_environment.test_filter", "name", env.Name),
					resource.TestCheckResourceAttr("data.honeycombio_environment.test_filter", "slug", env.Slug),
					resource.TestCheckResourceAttr("data.honeycombio_environment.test_filter", "color", *env.Color),
					resource.TestCheckResourceAttr("data.honeycombio_environment.test_filter", "description", *env.Description),
					resource.TestCheckResourceAttr("data.honeycombio_environment.test_filter", "delete_protected", "true"),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: `
data "honeycombio_environment" "test" {
  id = "boop"

  detail_filter {
    name  = "name"
    value = "boop"
  }
}`,
				ExpectError: regexp.MustCompile(`Attribute "detail_filter" cannot be specified when "id" is specified`),
				PlanOnly:    true,
			},
			{
				Config: `
data "honeycombio_environment" "test" {
  detail_filter {
    name  = "name"
    value = "boop"
  }
}`,
				ExpectError: regexp.MustCompile(`Your filter returned no matches`),
				PlanOnly:    true,
			},
		},
	})
}
