package honeycombio

import (
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAccDataSourceHoneycombioQuery_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryExists(t),
				),
			},
		},
	})
}

func testAccQueryConfig() string {
	return `
data "honeycombio_query" "test" {
    calculation {
        op     = "AVG"
        column = "duration_ms"
    }

    filter {
        column = "trace.parent_id"
        op     = "does-not-exist"
    }

    filter {
        column = "app.tenant"
        op     = "="
        value  = "ThatSpecialTenant"
    }
}

output "rendered" {
    value = data.honeycombio_query.test.rendered
}`
}

func testAccCheckQueryExists(t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rendered, ok := s.RootModule().Outputs["rendered"]
		if !ok {
			return errors.New("did not find output.rendered")
		}

		expectedRendered := `{
  "calculations": [
    {
      "op": "AVG",
      "column": "duration_ms"
    }
  ],
  "filters": [
    {
      "column": "trace.parent_id",
      "op": "does-not-exist",
      "value": ""
    },
    {
      "column": "app.tenant",
      "op": "=",
      "value": "ThatSpecialTenant"
    }
  ],
  "filter_combination": "AND"
}`

		ok = assert.Equal(t, expectedRendered, rendered.Value.(string))
		if !ok {
			return errors.New("rendered query did not match expected query")
		}

		return nil
	}
}
