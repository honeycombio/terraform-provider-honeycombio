package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	v2client "github.com/honeycombio/terraform-provider-honeycombio/client/v2"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAcc_EnvironmentsDatasource(t *testing.T) {
	ctx := context.Background()
	c := testAccV2Client(t)
	const numEnvs = 3
	// scope the regex filter to this test's environments so concurrently
	// running tests creating their own environments don't change the count
	envPrefix := test.RandomString(6) + "-"

	testEnvs := make([]*v2client.Environment, numEnvs)
	for i := range numEnvs {
		testEnvs[i] = testAccEnvironmentWithPrefix(ctx, t, c, envPrefix)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "honeycombio_environments" "all" {}

data "honeycombio_environments" "regex" {
  detail_filter {
    name        = "name"
    value_regex = "^test\\.%s"
  }
}

data "honeycombio_environments" "exact" {
  detail_filter {
    name  = "name"
    value = "%s"
  }
}`, envPrefix, testEnvs[0].Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.honeycombio_environments.regex",
						"ids.#",
						fmt.Sprintf("%d", numEnvs),
					),
					resource.TestCheckResourceAttr("data.honeycombio_environments.exact", "ids.#", "1"),
				),
			},
		},
	})
}
