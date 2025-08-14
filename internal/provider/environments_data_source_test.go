package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	v2client "github.com/honeycombio/terraform-provider-honeycombio/client/v2"
)

func TestAcc_EnvironmentsDatasource(t *testing.T) {
	ctx := t.Context()
	c := testAccV2Client(t)
	const numEnvs = 15

	// create a bunch of environments
	testEnvs := make([]*v2client.Environment, numEnvs)
	for i := range numEnvs {
		e := testAccEnvironment(ctx, t, c)
		testEnvs[i] = e
	}

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
}`, testEnvs[0].Name),
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
