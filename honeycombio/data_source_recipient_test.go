package honeycombio

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAccDataSourceHoneycombioRecipient_basic(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)

	testRecipients := []honeycombio.Recipient{
		{
			Type: honeycombio.RecipientTypeEmail,
			Details: honeycombio.RecipientDetails{
				EmailAddress: test.RandomEmail(),
			},
		},
		{
			Type: honeycombio.RecipientTypeEmail,
			Details: honeycombio.RecipientDetails{
				EmailAddress: test.RandomEmail(),
			},
		},
		{
			Type: honeycombio.RecipientTypeSlack,
			Details: honeycombio.RecipientDetails{
				SlackChannel: test.RandomStringWithPrefix("#test.", 12),
			},
		},
		{
			Type: honeycombio.RecipientTypeSlack,
			Details: honeycombio.RecipientDetails{
				SlackChannel: test.RandomStringWithPrefix("#test.", 12),
			},
		},
		{
			Type: honeycombio.RecipientTypePagerDuty,
			Details: honeycombio.RecipientDetails{
				PDIntegrationKey:  test.RandomString(32),
				PDIntegrationName: test.RandomStringWithPrefix("test.", 20),
			},
		},
		{
			Type: honeycombio.RecipientTypePagerDuty,
			Details: honeycombio.RecipientDetails{
				PDIntegrationKey:  test.RandomString(32),
				PDIntegrationName: test.RandomStringWithPrefix("test.", 20),
			},
		},
		{
			Type: honeycombio.RecipientTypeWebhook,
			Details: honeycombio.RecipientDetails{
				WebhookName:   test.RandomStringWithPrefix("test.", 16),
				WebhookSecret: test.RandomString(20),
				WebhookURL:    "https://my.webhook.dev.corp.io",
			},
		},
		{
			Type: honeycombio.RecipientTypeMSTeamsWorkflow,
			Details: honeycombio.RecipientDetails{
				WebhookName: test.RandomStringWithPrefix("test.", 16),
				WebhookURL:  "https://mycorp.westus.logic.azure.com/workflows/12345",
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
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccRecipientWithDeprecatedTarget("email", testRecipients[0].Details.EmailAddress),
				Check:  resource.TestCheckResourceAttr("data.honeycombio_recipient.test", "address", testRecipients[0].Details.EmailAddress),
			},
			{
				Config: testAccRecipientWithDeprecatedTarget("slack", testRecipients[2].Details.SlackChannel),
				Check:  resource.TestCheckResourceAttr("data.honeycombio_recipient.test", "channel", testRecipients[2].Details.SlackChannel),
			},
			{
				Config:      testAccRecipientWithDeprecatedTarget("email", "another@example.org"),
				ExpectError: regexp.MustCompile("your recipient query returned no results."),
			},
			{
				Config: testAccRecipientWithFilterValue("email", "address", testRecipients[1].Details.EmailAddress),
				Check:  resource.TestCheckResourceAttr("data.honeycombio_recipient.test", "address", testRecipients[1].Details.EmailAddress),
			},
			{
				Config:      testAccRecipientWithFilterValue("email", "address", "another@example.org"),
				ExpectError: regexp.MustCompile("your recipient query returned no results."),
			},
			{
				Config: testAccRecipientWithFilterValue("pagerduty", "integration_name", testRecipients[4].Details.PDIntegrationName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_recipient.test", "integration_name", testRecipients[4].Details.PDIntegrationName),
					resource.TestCheckResourceAttr("data.honeycombio_recipient.test", "integration_key", testRecipients[4].Details.PDIntegrationKey),
				),
			},
			{
				Config: testAccRecipientWithFilterRegex("webhook", "url", ".*dev.corp.io"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_recipient.test", "name", testRecipients[6].Details.WebhookName),
					resource.TestCheckResourceAttr("data.honeycombio_recipient.test", "secret", testRecipients[6].Details.WebhookSecret),
					resource.TestCheckResourceAttr("data.honeycombio_recipient.test", "url", "https://my.webhook.dev.corp.io"),
				),
			},
			{
				Config: testAccRecipientWithFilterValue("msteams", "name", testRecipients[7].Details.WebhookName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_recipient.test", "name", testRecipients[7].Details.WebhookName),
					resource.TestCheckResourceAttr("data.honeycombio_recipient.test", "url", "https://mycorp.westus.logic.azure.com/workflows/12345"),
				),
			},
			{
				Config: testAccRecipientWithFilterValue("slack", "channel", testRecipients[3].Details.SlackChannel),
				Check:  resource.TestCheckResourceAttr("data.honeycombio_recipient.test", "channel", testRecipients[3].Details.SlackChannel),
			},
			{
				Config:      testAccRecipientWithFilterRegex("email", "address", ".*@example.com"),
				ExpectError: regexp.MustCompile("your recipient query returned more than one result. Please try a more specific search criteria."),
			},
			{
				Config:      testAccRecipientWithFilterRegex("pagerduty", "integration_name", ".*"),
				ExpectError: regexp.MustCompile("your recipient query returned more than one result. Please try a more specific search criteria."),
			},
		},
	})
}

func testAccRecipientWithDeprecatedTarget(recipientType, target string) string {
	return fmt.Sprintf(`
data "honeycombio_recipient" "test" {
  type   = "%s"
  target = "%s"
}`, recipientType, target)
}

func testAccRecipientWithFilterValue(recipientType, filterName, filterValue string) string {
	return fmt.Sprintf(`
data "honeycombio_recipient" "test" {
  type   = "%s"

  detail_filter {
    name  = "%s"
    value = "%s"
  }
}`, recipientType, filterName, filterValue)
}

func testAccRecipientWithFilterRegex(recipientType, filterName, filterRegex string) string {
	return fmt.Sprintf(`
data "honeycombio_recipient" "test" {
  type   = "%s"

  detail_filter {
    name        = "%s"
    value_regex = "%s"
  }
}`, recipientType, filterName, filterRegex)
}
