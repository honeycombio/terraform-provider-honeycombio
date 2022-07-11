package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
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

		rcpt, err = c.Recipients.Create(ctx, data)

		assert.NoError(t, err, "unable to create Recipient")
		assert.NotNil(t, rcpt.ID, "ID is empty")
		assert.NotNil(t, rcpt.CreatedAt, "created at is empty")
		assert.NotNil(t, rcpt.UpdatedAt, "updated at is empty")
		// copy dynamic fields before asserting equality
		data.ID = rcpt.ID
		data.CreatedAt = rcpt.CreatedAt
		data.UpdatedAt = rcpt.UpdatedAt
		assert.Equal(t, data, rcpt)
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

		result, err := c.Recipients.Update(ctx, rcpt)

		assert.NoError(t, err, "failed to update Recipient")
		// copy dynamic field before asserting equality
		rcpt.UpdatedAt = result.UpdatedAt
		assert.Equal(t, result, rcpt)
	})

	t.Run("Delete", func(t *testing.T) {
		err = c.Recipients.Delete(ctx, rcpt.ID)

		assert.NoError(t, err, "failed to delete Recipient")
	})

	t.Run("Get_NotFound", func(t *testing.T) {
		_, err := c.Recipients.Get(ctx, rcpt.ID)

		assert.Equal(t, ErrNotFound, err)
	})
}
