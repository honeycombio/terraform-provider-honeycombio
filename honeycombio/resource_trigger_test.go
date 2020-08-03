package honeycombio

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	honeycombio "github.com/kvrhdn/go-honeycombio"
	"github.com/stretchr/testify/assert"
)

func TestAccHoneycombioTrigger_basic(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig(dataset),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(t, "honeycombio_trigger.test"),
				),
			},
		},
	})
}

func testAccTriggerConfig(dataset string) string {
	return fmt.Sprintf(`
data "honeycombio_query" "test" {
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }
}

resource "honeycombio_trigger" "test" {
  name    = "Test trigger from terraform-provider-honeycombio"
  dataset = "%s"

  query_json = data.honeycombio_query.test.json
  
  threshold {
    op    = ">"
    value = 100
  }
  
  recipient {
    type   = "email"
    target = "hello@example.com"
  }
  
  recipient {
    type   = "email"
    target = "bye@example.com"
  }
}`, dataset)
}

func testAccCheckTriggerExists(t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		client := testAccProvider.Meta().(*honeycombio.Client)
		createdTrigger, err := client.Triggers.Get(resourceState.Primary.Attributes["dataset"], resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("could not find created trigger: %w", err)
		}

		expectedTrigger := &honeycombio.Trigger{
			ID:          createdTrigger.ID,
			Name:        "Test trigger from terraform-provider-honeycombio",
			Description: "",
			Disabled:    false,
			Query: &honeycombio.QuerySpec{
				Calculations: []honeycombio.CalculationSpec{
					{
						Op:     honeycombio.CalculateOpAvg,
						Column: &[]string{"duration_ms"}[0],
					},
				},
			},
			Frequency: 900,
			Threshold: &honeycombio.TriggerThreshold{
				Op:    honeycombio.TriggerThresholdOpGreaterThan,
				Value: &[]float64{100}[0],
			},
			Recipients: []honeycombio.TriggerRecipient{
				{
					Type:   "email",
					Target: "hello@example.com",
				},
				{
					Type:   "email",
					Target: "bye@example.com",
				},
			},
		}

		ok = assert.Equal(t, expectedTrigger, createdTrigger)
		if !ok {
			return errors.New("created trigger did not match expected trigger")
		}
		return nil
	}
}

func TestAccHoneycombioTrigger_validationErrors(t *testing.T) {
	dataset := testAccDataset()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTriggerConfigWithQuery(dataset, `{]`),
				ExpectError: regexp.MustCompile("Value of query_json is not a valid query specification"),
			},
			{
				Config:      testAccTriggerConfigWithQuery(dataset, `{"calculations":"bar"}`),
				ExpectError: regexp.MustCompile("Value of query_json is not a valid query specification"),
			},
			{
				Config: testAccTriggerConfigWithQuery(dataset, `
{
    "calculations": [
        {"op": "COUNT"},
        {"op": "AVG", "column": "duration_ms"}
    ]
}`),
				ExpectError: regexp.MustCompile("Query of a trigger must have exactly one calculation"),
			},
		},
	})
}

func testAccTriggerConfigWithQuery(dataset, query string) string {
	return fmt.Sprintf(`
resource "honeycombio_trigger" "test" {
  name    = "Test trigger from terraform-provider-honeycombio"
  dataset = "%s"

  query_json = <<EOF
%s
EOF
    
  threshold {
    op    = ">"
    value = 100
  }
}`, dataset, query)
}
