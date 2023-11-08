package honeycombio

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestAccDataSourceHoneycombioRecipients_basic(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)

	testRecipients := []honeycombio.Recipient{
		{
			Type: honeycombio.RecipientTypeEmail,
			Details: honeycombio.RecipientDetails{
				EmailAddress: "acctest@example.net",
			},
		},
		{
			Type: honeycombio.RecipientTypeEmail,
			Details: honeycombio.RecipientDetails{
				EmailAddress: "acctest2@example.net",
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
				Config: `data "honeycombio_recipients" "all" {}`,
			},
			{
				Config: `data "honeycombio_recipients" "email" { type = "email" }`,
			},
			{
				Config: testAccRecipientsWithFilterValue("email", "address", "acctest@example.net"),
				Check:  resource.TestCheckResourceAttr("data.honeycombio_recipients.test", "ids.#", "1"),
			},
			{
				Config: testAccRecipientsWithFilterValue("email", "address", "another@example.net"),
				Check:  resource.TestCheckResourceAttr("data.honeycombio_recipients.test", "ids.#", "0"),
			},
			{
				Config: testAccRecipientsWithFilterRegex("slack", "channel", "^#.*acctest"),
				Check:  resource.TestCheckResourceAttr("data.honeycombio_recipients.test", "ids.#", "2"),
			},
		},
	})
}

func testAccRecipientsWithFilterValue(recipientType, filterName, filterValue string) string {
	return fmt.Sprintf(`
data "honeycombio_recipients" "test" {
  type   = "%s"

  detail_filter {
    name  = "%s"
    value = "%s"
  }
}
`, recipientType, filterName, filterValue)
}

func testAccRecipientsWithFilterRegex(recipientType, filterName, filterRegex string) string {
	return fmt.Sprintf(`
data "honeycombio_recipients" "test" {
  type   = "%s"

  detail_filter {
    name        = "%s"
    value_regex = "%s"
  }
}
`, recipientType, filterName, filterRegex)
}
