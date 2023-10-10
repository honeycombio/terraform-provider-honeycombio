package client

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecipientsEmail(t *testing.T) {
	ctx := context.Background()

	var rcpt *Recipient
	var err error

	c := newTestClient(t)

	t.Run("Create", func(t *testing.T) {
		data := &Recipient{
			Type: RecipientTypeEmail,
			Details: RecipientDetails{
				EmailAddress: "hnytest@example.com",
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

		assert.NoError(t, err, "unable to list Recipients")
		assert.Contains(t, results, *rcpt, "could not find newly created Recipient with List")
	})

	t.Run("Get", func(t *testing.T) {
		result, err := c.Recipients.Get(ctx, rcpt.ID)

		assert.NoError(t, err, "failed to get Recipient by ID")
		assert.Equal(t, *rcpt, *result)
	})

	t.Run("Update", func(t *testing.T) {
		rcpt.Details.EmailAddress = "hnytest2@example.com"
		now := time.Now()
		result, err := c.Recipients.Update(ctx, rcpt)

		assert.NoError(t, err, "failed to update Recipient")
		assert.Equal(t, rcpt.Details.EmailAddress, result.Details.EmailAddress, "email address not updated")
		assert.WithinDuration(t, now, result.UpdatedAt, 2*time.Second)
	})

	t.Run("Delete", func(t *testing.T) {
		err = c.Recipients.Delete(ctx, rcpt.ID)
		assert.NoError(t, err, "failed to delete Recipient")
	})

	t.Run("Fail to Get deleted Recipient", func(t *testing.T) {
		_, err := c.Recipients.Get(ctx, rcpt.ID)

		var de DetailedError
		assert.Error(t, err)
		assert.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())
	})
}
