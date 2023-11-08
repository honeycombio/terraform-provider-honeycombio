package honeycombio

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

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
		{
			Type: honeycombio.RecipientTypeMSTeams,
			Details: honeycombio.RecipientDetails{
				WebhookName: "My Teams Channel",
				WebhookURL:  "https://outlook.office.com/webhook/12345",
			},
		},
	}

	for i, r := range testRecipients {
		rcpt, err := c.Recipients.Create(ctx, &r)
		require.NoError(t, err)
		// update ID for removal later
		testRecipients[i].ID = rcpt.ID
	}
	//nolint:errcheck
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
				Check:  resource.TestCheckResourceAttr("data.honeycombio_recipient.test", "address", "acctest@example.org"),
			},
			{
				Config: testAccRecipientWithDeprecatedTarget("slack", "#acctest"),
				Check:  resource.TestCheckResourceAttr("data.honeycombio_recipient.test", "channel", "#acctest"),
			},
			{
				Config:      testAccRecipientWithDeprecatedTarget("email", "another@example.org"),
				ExpectError: regexp.MustCompile("your recipient query returned no results."),
			},
			{
				Config: testAccRecipientWithFilterValue("email", "address", "acctest2@example.org"),
				Check:  resource.TestCheckResourceAttr("data.honeycombio_recipient.test", "address", "acctest2@example.org"),
			},
			{
				Config:      testAccRecipientWithFilterValue("email", "address", "another@example.org"),
				ExpectError: regexp.MustCompile("your recipient query returned no results."),
			},
			{
				Config: testAccRecipientWithFilterValue("pagerduty", "integration_name", "My Important Service"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_recipient.test", "integration_name", "My Important Service"),
					resource.TestCheckResourceAttr("data.honeycombio_recipient.test", "integration_key", "6f05176bf1c7a1adb6ee516521770ec4"),
				),
			},
			{
				Config: testAccRecipientWithFilterRegex("webhook", "url", ".*dev.corp.io"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_recipient.test", "name", "My Notifications Hook"),
					resource.TestCheckResourceAttr("data.honeycombio_recipient.test", "secret", "s0s3kret!"),
					resource.TestCheckResourceAttr("data.honeycombio_recipient.test", "url", "https://my.webhook.dev.corp.io"),
				),
			},
			{
				Config: testAccRecipientWithFilterValue("msteams", "name", "My Teams Channel"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_recipient.test", "name", "My Teams Channel"),
					resource.TestCheckResourceAttr("data.honeycombio_recipient.test", "url", "https://outlook.office.com/webhook/12345"),
				),
			},
			{
				Config: testAccRecipientWithFilterValue("slack", "channel", "#tmp-acctest"),
				Check:  resource.TestCheckResourceAttr("data.honeycombio_recipient.test", "channel", "#tmp-acctest"),
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
data "honeycombio_recipient" "test" {
  type   = "%s"
  target = "%s"
}
`, recipientType, target)
}

func testAccRecipientWithFilterValue(recipientType, filterName, filterValue string) string {
	return fmt.Sprintf(`
data "honeycombio_recipient" "test" {
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
data "honeycombio_recipient" "test" {
  type   = "%s"

  detail_filter {
    name        = "%s"
    value_regex = "%s"
  }
}
`, recipientType, filterName, filterRegex)
}
