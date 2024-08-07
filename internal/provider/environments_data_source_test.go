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
	const numEnvs = 15

	// create a bunch of environments
	testEnvs := make([]*v2client.Environment, numEnvs+1)
	for i := 0; i < numEnvs; i++ {
		e := testAccEnvironment(ctx, t, c)
		testEnvs[i] = e
	}
	// one additional with a different prefix for filter testing
	e := testAccEnvironment(ctx, t, c)
	testEnvs[numEnvs] = e

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "honeycombio_environments" "all" {}

data "honeycombio_environments" "regex" {
  detail_filter {
    name        = "name"
    value_regex = "test.*"
  }
}

data "honeycombio_environments" "exact" {
  detail_filter {
    name  = "name"
    value = "%s"
  }
}`, e.Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					test.TestCheckResourceAttrAtLeast(
						"data.honeycombio_environments.all",
						"ids.#",
						numEnvs+2), // +2 because of the additional environment created and the 'ci' environment
					resource.TestCheckResourceAttr(
						"data.honeycombio_environments.regex",
						"ids.#",
						fmt.Sprintf("%d", numEnvs+1), // +1 because of the additional environment created
					),
					resource.TestCheckResourceAttr("data.honeycombio_environments.exact", "ids.#", "1"),
				),
			},
		},
	})
}
