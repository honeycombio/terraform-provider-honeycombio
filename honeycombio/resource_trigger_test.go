package honeycombio

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestAccHoneycombioTrigger_basic(t *testing.T) {
	var triggerBefore, triggerAfter honeycombio.Trigger

	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		IDRefreshName:     "honeycombio_trigger.test",
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfigWithFrequency(dataset, 900),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(t, "honeycombio_trigger.test", &triggerBefore),
					testAccCheckTriggerAttributes(&triggerBefore),
					resource.TestCheckResourceAttr("honeycombio_trigger.test", "frequency", "900"),
					resource.TestCheckResourceAttr("honeycombio_trigger.test", "frequency", "900"),
					resource.TestCheckResourceAttr("honeycombio_trigger.test", "alert_type", "on_change"),
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
			{
				ResourceName:        "honeycombio_trigger.test",
				ImportStateIdPrefix: fmt.Sprintf("%v/", dataset),
				ImportState:         true,
				ImportStateVerify:   true,
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

		client := testAccClient(t)
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
			return fmt.Errorf("bad name: %s", t.Name)
		}

		if t.Frequency != 900 {
			return fmt.Errorf("bad frequency: %d", t.Frequency)
		}

		if t.AlertType != "on_change" {
			return fmt.Errorf("bad AlertType: %s | should be on_change", t.AlertType)
		}

		return nil
	}
}

// add a trigger recipient by ID to verify the diff is stable
func TestAccHoneycombioTrigger_triggerRecipientById(t *testing.T) {
	dataset := testAccDataset()

	trigger, deleteFn := createTriggerWithRecipient(t, dataset, honeycombio.TriggerRecipient{
		Type:   honeycombio.TriggerRecipientTypeEmail,
		Target: "acctest@example.com",
	})
	defer deleteFn()

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		IDRefreshName:     "honeycombio_trigger.test",
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfigWithRecipientID(dataset, trigger.Recipients[0].ID),
			},
		},
	})
}

func TestAccHoneycombioTrigger_recipientOrderingNoDiff(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "honeycombio_query_specification" "test" {
  calculation {
    op     = "COUNT"
  }
}

resource "honeycombio_query" "test" {
  dataset    = "%s"
  query_json = data.honeycombio_query_specification.test.json
}

resource "honeycombio_trigger" "test" {
  name = "Ensure mixed order recipients don't cause infinite diffs"

  query_id = honeycombio_query.test.id
  dataset  = "%s"

  threshold {
    op    = ">"
    value = 1000
  }

  alert_type = "on_change"

  recipient {
    type   = "slack"
    target = "#test2"
  }

  recipient {
    type   = "slack"
    target = "#test"
  }

  recipient {
    type   = "email"
    target = "bob@example.com"
  }

  recipient {
    type   = "email"
    target = "alice@example.com"
  }

  recipient {
    type   = "marker"
    target = "trigger fired"
  }
}
`, dataset, dataset),
			},
		},
	})

}

func testAccTriggerConfigWithFrequency(dataset string, frequency int) string {
	return fmt.Sprintf(`
data "honeycombio_query_specification" "test" {
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }
  time_range = 1200
}

resource "honeycombio_query" "test" {
  dataset    = "%s"
  query_json = data.honeycombio_query_specification.test.json
}

resource "honeycombio_trigger" "test" {
  name    = "Test trigger from terraform-provider-honeycombio"
  dataset = "%s"

  query_id = honeycombio_query.test.id

  alert_type = "on_change"
  
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
}`, dataset, dataset, frequency)
}

func testAccTriggerConfigWithCount(dataset string) string {
	return fmt.Sprintf(`
data "honeycombio_query_specification" "test" {
  time_range = 1200

  calculation {
    op     = "COUNT"
  }

  filter_combination = "AND"

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

resource "honeycombio_query" "test" {
  dataset    = "%s"
  query_json = data.honeycombio_query_specification.test.json
}

resource "honeycombio_trigger" "test" {
  name    = "Test trigger from terraform-provider-honeycombio"
  dataset = "%s"

  query_id = honeycombio_query.test.id

  alert_type = "on_change"

  threshold {
    op    = ">"
    value = 100
  }
}`, dataset, dataset)
}

func testAccTriggerConfigWithRecipientID(dataset, recipientID string) string {
	return fmt.Sprintf(`
data "honeycombio_query_specification" "test" {
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }
  time_range = 1800
}

resource "honeycombio_query" "test" {
  dataset    = "%s"
  query_json = data.honeycombio_query_specification.test.json
}

resource "honeycombio_trigger" "test" {
  name    = "Test trigger from terraform-provider-honeycombio"
  dataset = "%s"

  query_id = honeycombio_query.test.id

  alert_type = "on_change"

  threshold {
    op    = ">"
    value = 100
  }

  recipient {
    id = "%s"
  }
}`, dataset, dataset, recipientID)
}
