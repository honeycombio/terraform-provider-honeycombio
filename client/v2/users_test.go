package v2

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Users(t *testing.T) {
	ctx := context.Background()

	c, f := newFakeAPIClient(t, map[string]http.HandlerFunc{
		"GET /2/teams/" + testTeamSlug + "/users": func(w http.ResponseWriter, _ *http.Request) {
			writeJSONAPI(w, http.StatusOK, `{"data": [{"type": "users", "id": "hcxus_a",
				"attributes": {"email": "alex@example.com", "role": "member"}}], "links": {}}`)
		},
	})

	pager, err := c.Users.List(ctx, Filter("email", "Alex+ops@Example.com"))
	require.NoError(t, err)
	users, err := pager.Next(ctx)
	require.NoError(t, err)

	reqs := f.recorded()
	require.Len(t, reqs, 1)
	assert.Equal(t, "filter%5Bemail%5D=Alex%2Bops%40Example.com&page%5Bsize%5D=20", reqs[0].Query)
	require.Len(t, users, 1)
	assert.Equal(t, User{ID: "hcxus_a", Email: "alex@example.com", Role: "member"}, *users[0])
}
