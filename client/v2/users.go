package v2

import (
	"context"
	"fmt"
)

// Compile-time proof of interface implementation.
var _ Users = (*users)(nil)

// Users lists the members of the Team.
type Users interface {
	// List lists the Team's users. Pending invites are not listed.
	//
	// It accepts Filter("email", address), which matches without regard to
	// case, and Filter("group", groupID), which needs the groups:read scope.
	List(ctx context.Context, opts ...ListOption) (*Pager[User], error)
}

type users struct {
	client *Client
}

const usersPath = "/2/teams/%s/users"

func (u *users) List(ctx context.Context, opts ...ListOption) (*Pager[User], error) {
	slug, err := u.client.teamSlug(ctx)
	if err != nil {
		return nil, err
	}
	return NewPager[User](u.client, fmt.Sprintf(usersPath, slug), opts...)
}
