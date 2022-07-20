package honeycombio

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestAccDataSourceHoneycombioRecipient_basic(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)

	testRecipients := []honeycombio.Recipient{
		{
			Type: honeycombio.RecipientTypeEmail,
			Details: honeycombio.RecipientDetails{
				EmailAddress: "acctest@example.org",
			},
		},
		{
			Type: honeycombio.RecipientTypeEmail,
			Details: honeycombio.RecipientDetails{
				EmailAddress: "acctest2@example.org",
			},
		},
		{
			Type: honeycombio.RecipientTypeSlack,
			Details: honeycombio.RecipientDetails{
				SlackChannel: "#acctest",
			},
		},
		{
			Type: honeycombio.RecipientTypeSlack,
			Details: honeycombio.RecipientDetails{
				SlackChannel: "#tmp-acctest",
			},
		},
		{
			Type: honeycombio.RecipientTypePagerDuty,
			Details: honeycombio.RecipientDetails{
				PDIntegrationKey:  "6f05176bf1c7a1adb6ee516521770ec4",
				PDIntegrationName: "My Important Service",
			},
		},
		{
			Type: honeycombio.RecipientTypePagerDuty,
			Details: honeycombio.RecipientDetails{
				PDIntegrationKey:  "6f05176bf1b7a1adb6ee516521770ac0",
				PDIntegrationName: "My Other Important Service",
			},
		},
		{
			Type: honeycombio.RecipientTypeWebhook,
			Details: honeycombio.RecipientDetails{
				WebhookName:   "My Notifications Hook",
				WebhookSecret: "s0s3kret!",
				WebhookURL:    "https://my.webhook.dev.corp.io",
			},
		},
	}

	for i, r := range testRecipients {
		rcpt, err := c.Recipients.Create(ctx, &r)
		// update ID for removal later
		testRecipients[i].ID = rcpt.ID
		if err != nil {
			t.Error(err)
		}
	}
	t.Cleanup(func() {
		// remove Recipients at the of the test run
		for _, r := range testRecipients {
			c.Recipients.Delete(ctx, r.ID)
		}
	})

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRecipientWithDeprecatedTarget("email", "acctest@example.org"),
			},
			{
				Config: testAccRecipientWithDeprecatedTarget("slack", "#acctest"),
			},
			{
				Config:      testAccRecipientWithDeprecatedTarget("email", "another@example.org"),
				ExpectError: regexp.MustCompile("your recipient query returned no results."),
			},
			{
				Config: testAccRecipientWithFilterValue("email", "address", "acctest2@example.org"),
			},
			{
				Config:      testAccRecipientWithFilterValue("email", "address", "another@example.org"),
				ExpectError: regexp.MustCompile("your recipient query returned no results."),
			},
			{
				Config: testAccRecipientWithFilterRegex("webhook", "url", ".*dev.corp.io"),
			},
			{
				Config: testAccRecipientWithFilterValue("slack", "channel", "#tmp-acctest"),
			},
			{
				Config:      testAccRecipientWithFilterRegex("email", "address", "^acctest*"),
				ExpectError: regexp.MustCompile("your recipient query returned more than one result. Please try a more specific search critera."),
			},
			{
				Config:      testAccRecipientWithFilterRegex("pagerduty", "integration_name", "^.*Important Service$"),
				ExpectError: regexp.MustCompile("your recipient query returned more than one result. Please try a more specific search critera."),
			},
		},
	})
}

func testAccRecipientWithDeprecatedTarget(recipientType, target string) string {
	return fmt.Sprintf(`
data "honeycombio_recipient" "test_target" {
  type   = "%s"
  target = "%s"
}
`, recipientType, target)
}

func testAccRecipientWithFilterValue(recipientType, filterName, filterValue string) string {
	return fmt.Sprintf(`
data "honeycombio_recipient" "test_filter_value" {
  type   = "%s"

  detail_filter {
    name  = "%s"
    value = "%s"
  }
}
`, recipientType, filterName, filterValue)
}

func testAccRecipientWithFilterRegex(recipientType, filterName, filterRegex string) string {
	return fmt.Sprintf(`
data "honeycombio_recipient" "test_filter_regex" {
  type   = "%s"

  detail_filter {
    name        = "%s"
    value_regex = "%s"
  }
}
`, recipientType, filterName, filterRegex)
}
