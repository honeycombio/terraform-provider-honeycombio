package client_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestRecipientsEmail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var rcpt *client.Recipient
	var err error

	c := newTestClient(t)

	t.Run("Create", func(t *testing.T) {
		data := &client.Recipient{
			Type: client.RecipientTypeEmail,
			Details: client.RecipientDetails{
				EmailAddress: test.RandomEmail(),
			},
		}
		now := time.Now()
		rcpt, err = c.Recipients.Create(ctx, data)

		require.NoError(t, err, "unable to create Recipient")
		assert.NotNil(t, rcpt.ID, "ID is empty")
		assert.NotNil(t, rcpt.CreatedAt, "created at is empty")
		assert.NotNil(t, rcpt.UpdatedAt, "updated at is empty")
		assert.Equal(t, data.Details.EmailAddress, rcpt.Details.EmailAddress, "email address does not match")
		assert.WithinDuration(t, now, rcpt.CreatedAt, 2*time.Second)
		assert.WithinDuration(t, now, rcpt.UpdatedAt, 2*time.Second)
	})

	t.Run("List", func(t *testing.T) {
		results, err := c.Recipients.List(ctx)

		require.NoError(t, err, "unable to list Recipients")
		assert.Contains(t, results, *rcpt, "could not find newly created Recipient with List")
	})

	t.Run("Get", func(t *testing.T) {
		result, err := c.Recipients.Get(ctx, rcpt.ID)

		require.NoError(t, err, "failed to get Recipient by ID")
		assert.Equal(t, *rcpt, *result)
	})

	t.Run("Update", func(t *testing.T) {
		rcpt.Details.EmailAddress = test.RandomEmail()
		now := time.Now()
		result, err := c.Recipients.Update(ctx, rcpt)

		require.NoError(t, err, "failed to update Recipient")
		assert.Equal(t, rcpt.Details.EmailAddress, result.Details.EmailAddress, "email address not updated")
		assert.WithinDuration(t, now, result.UpdatedAt, 2*time.Second)
	})

	t.Run("Delete", func(t *testing.T) {
		err = c.Recipients.Delete(ctx, rcpt.ID)
		require.NoError(t, err, "failed to delete Recipient")
	})

	t.Run("Fail to Get deleted Recipient", func(t *testing.T) {
		_, err := c.Recipients.Get(ctx, rcpt.ID)

		var de client.DetailedError
		require.Error(t, err)
		require.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())
	})
}

func TestRecipientsWebhooksandMSTeams(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	c := newTestClient(t)

	testCases := []struct {
		rcpt      client.Recipient
		expectErr bool
	}{
		{
			rcpt: client.Recipient{
				Type: client.RecipientTypeWebhook,
				Details: client.RecipientDetails{
					WebhookName:   test.RandomStringWithPrefix("test.", 10),
					WebhookURL:    test.RandomURL(),
					WebhookSecret: "secret",
				},
			},
		},
		{
			rcpt: client.Recipient{
				Type: client.RecipientTypeMSTeams,
				Details: client.RecipientDetails{
					WebhookName: test.RandomStringWithPrefix("test.", 10),
					WebhookURL:  test.RandomURL(),
				},
			},
			expectErr: true, // creation of new MSTeams recipients is not allowed
		},
		{
			rcpt: client.Recipient{
				Type: client.RecipientTypeMSTeamsWorkflow,
				Details: client.RecipientDetails{
					WebhookName: test.RandomStringWithPrefix("test.", 10),
					WebhookURL:  test.RandomURL(),
				},
			},
		},
	}

	for _, tc := range testCases {
		tr := tc.rcpt
		t.Run(tr.Type.String(), func(t *testing.T) {
			r, err := c.Recipients.Create(ctx, &tr)
			t.Cleanup(func() {
				_ = c.Recipients.Delete(ctx, r.ID)
			})

			if tc.expectErr {
				require.Error(t, err, "expected error creating %s recipient", tr.Type)
				return
			}
			require.NoError(t, err, "failed to create %s recipient", tr.Type)
			r, err = c.Recipients.Get(ctx, r.ID)
			require.NoError(t, err)

			assert.Equal(t, tr.Type, r.Type)
			assert.Equal(t, tr.Details.WebhookName, r.Details.WebhookName)
			assert.Equal(t, tr.Details.WebhookURL, r.Details.WebhookURL)
			assert.Equal(t, tr.Details.WebhookSecret, r.Details.WebhookSecret)
		})
	}
}
