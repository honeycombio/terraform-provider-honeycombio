package honeycombio

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	honeycombio "github.com/kvrhdn/go-honeycombio"
)

func TestAccHoneycombioTrigger_basic(t *testing.T) {
	var triggerBefore, triggerAfter honeycombio.Trigger

	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "honeycombio_trigger.test",
		Providers:     testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfigWithFrequency(dataset, 900),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(t, "honeycombio_trigger.test", &triggerBefore),
					testAccCheckTriggerAttributes(&triggerBefore),
					resource.TestCheckResourceAttr("honeycombio_trigger.test", "frequency", "900"),
				),
			},
			{
				Config: testAccTriggerConfigWithFrequency(dataset, 300),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(t, "honeycombio_trigger.test", &triggerAfter),
					resource.TestCheckResourceAttr("honeycombio_trigger.test", "frequency", "300"),
				),
			},
			{
				Config: testAccTriggerConfigWithCount(dataset),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(t, "honeycombio_trigger.test", &triggerAfter),
				),
			},
		},
	})
}

func testAccCheckTriggerExists(t *testing.T, name string, trigger *honeycombio.Trigger) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		client := testAccProvider.Meta().(*honeycombio.Client)
		createdTrigger, err := client.Triggers.Get(context.Background(), resourceState.Primary.Attributes["dataset"], resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("could not find created trigger: %w", err)
		}

		*trigger = *createdTrigger
		return nil
	}
}

func testAccCheckTriggerAttributes(t *honeycombio.Trigger) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if t.Name != "Test trigger from terraform-provider-honeycombio" {
			return fmt.Errorf("Bad name: %s", t.Name)
		}

		if t.Frequency != 900 {
			return fmt.Errorf("Bad frequency: %d", t.Frequency)
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

func testAccTriggerConfigWithFrequency(dataset string, frequency int) string {
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

  frequency = %d

  recipient {
    type   = "email"
    target = "hello@example.com"
  }

  recipient {
    type   = "email"
    target = "bye@example.com"
  }
}`, dataset, frequency)
}

func testAccTriggerConfigWithCount(dataset string) string {
	return fmt.Sprintf(`
data "honeycombio_query" "test" {
  calculation {
    op     = "COUNT"
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
}`, dataset)
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
