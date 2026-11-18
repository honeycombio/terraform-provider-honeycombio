package v2

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/hashicorp/jsonapi"

	hnyclient "github.com/honeycombio/terraform-provider-honeycombio/client"
)

// Compile-time proof of interface implementation.
var _ Groups = (*groups)(nil)

// Groups manages a Team's Groups, their members and their dataset grants.
type Groups interface {
	Create(ctx context.Context, g *Group) (*Group, error)
	Get(ctx context.Context, id string) (*Group, error)
	// Update changes the Group's name and description.
	Update(ctx context.Context, g *Group) (*Group, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, opts ...ListOption) (*Pager[Group], error)

	// ListMembers lists the users in the Group. Agent members are never listed.
	ListMembers(ctx context.Context, groupID string, opts ...ListOption) (*Pager[GroupMember], error)
	// ReplaceMembers makes the Group's users exactly userIDs, in one change.
	// An empty list removes every user. Agent members are left unchanged.
	ReplaceMembers(ctx context.Context, groupID string, userIDs []string) error

	// ListDatasets lists the datasets the Group is granted.
	ListDatasets(ctx context.Context, groupID string, opts ...ListOption) (*Pager[GroupDataset], error)
	// GrantDatasets grants the Group access to datasetIDs. Granting a dataset
	// the Group already has changes nothing.
	GrantDatasets(ctx context.Context, groupID string, datasetIDs []string) error
	// RevokeDatasets revokes the Group's access to datasetIDs. Revoking a
	// dataset the Group does not have changes nothing.
	RevokeDatasets(ctx context.Context, groupID string, datasetIDs []string) error
}

type groups struct {
	client *Client
}

const (
	groupsPath        = "/2/teams/%s/groups"
	groupsByIDPath    = "/2/teams/%s/groups/%s"
	groupMembersPath  = "/2/teams/%s/groups/%s/relationships/members"
	groupDatasetsPath = "/2/teams/%s/groups/%s/relationships/datasets"

	// maxDatasetsPerGrantRequest is the largest batch the grant and revoke
	// routes accept.
	maxDatasetsPerGrantRequest = 100
)

func (g *groups) Create(ctx context.Context, group *Group) (*Group, error) {
	slug, err := g.client.teamSlug(ctx)
	if err != nil {
		return nil, err
	}
	return g.write(ctx, http.MethodPost, fmt.Sprintf(groupsPath, slug), http.StatusCreated, group)
}

func (g *groups) Get(ctx context.Context, id string) (*Group, error) {
	slug, err := g.client.teamSlug(ctx)
	if err != nil {
		return nil, err
	}
	r, err := g.client.Do(ctx, http.MethodGet, fmt.Sprintf(groupsByIDPath, slug, id), nil)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, hnyclient.ErrorFromResponse(r)
	}

	group := new(Group)
	if err := jsonapi.UnmarshalPayload(r.Body, group); err != nil {
		return nil, err
	}
	return group, nil
}

func (g *groups) Update(ctx context.Context, group *Group) (*Group, error) {
	slug, err := g.client.teamSlug(ctx)
	if err != nil {
		return nil, err
	}
	return g.write(ctx, http.MethodPatch, fmt.Sprintf(groupsByIDPath, slug, group.ID), http.StatusOK, group)
}

// write sends group and returns the Group in the response.
func (g *groups) write(ctx context.Context, method, url string, wantStatus int, group *Group) (*Group, error) {
	r, err := g.client.Do(ctx, method, url, group)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != wantStatus {
		return nil, hnyclient.ErrorFromResponse(r)
	}

	result := new(Group)
	if err := jsonapi.UnmarshalPayload(r.Body, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (g *groups) Delete(ctx context.Context, id string) error {
	slug, err := g.client.teamSlug(ctx)
	if err != nil {
		return err
	}
	return g.send(ctx, http.MethodDelete, fmt.Sprintf(groupsByIDPath, slug, id), nil)
}

func (g *groups) List(ctx context.Context, opts ...ListOption) (*Pager[Group], error) {
	slug, err := g.client.teamSlug(ctx)
	if err != nil {
		return nil, err
	}
	return NewPager[Group](g.client, fmt.Sprintf(groupsPath, slug), opts...)
}

func (g *groups) ListMembers(ctx context.Context, groupID string, opts ...ListOption) (*Pager[GroupMember], error) {
	slug, err := g.client.teamSlug(ctx)
	if err != nil {
		return nil, err
	}
	return NewPager[GroupMember](g.client, fmt.Sprintf(groupMembersPath, slug, groupID), opts...)
}

func (g *groups) ReplaceMembers(ctx context.Context, groupID string, userIDs []string) error {
	slug, err := g.client.teamSlug(ctx)
	if err != nil {
		return err
	}
	members := make([]*GroupMember, len(userIDs))
	for i, id := range userIDs {
		members[i] = &GroupMember{ID: id}
	}
	err = g.send(ctx, http.MethodPatch, fmt.Sprintf(groupMembersPath, slug, groupID), members)
	return withRequestIDs(err, userIDs)
}

func (g *groups) ListDatasets(ctx context.Context, groupID string, opts ...ListOption) (*Pager[GroupDataset], error) {
	slug, err := g.client.teamSlug(ctx)
	if err != nil {
		return nil, err
	}
	return NewPager[GroupDataset](g.client, fmt.Sprintf(groupDatasetsPath, slug, groupID), opts...)
}

func (g *groups) GrantDatasets(ctx context.Context, groupID string, datasetIDs []string) error {
	return g.changeDatasets(ctx, http.MethodPost, groupID, datasetIDs)
}

func (g *groups) RevokeDatasets(ctx context.Context, groupID string, datasetIDs []string) error {
	return g.changeDatasets(ctx, http.MethodDelete, groupID, datasetIDs)
}

// changeDatasets sends datasetIDs in batches the API accepts. Each batch
// is all or nothing, but the batches are not: when one fails, the earlier
// batches stay applied.
func (g *groups) changeDatasets(ctx context.Context, method, groupID string, datasetIDs []string) error {
	slug, err := g.client.teamSlug(ctx)
	if err != nil {
		return err
	}
	url := fmt.Sprintf(groupDatasetsPath, slug, groupID)
	for start := 0; start < len(datasetIDs); start += maxDatasetsPerGrantRequest {
		end := min(start+maxDatasetsPerGrantRequest, len(datasetIDs))
		batch := make([]*GroupDataset, 0, end-start)
		for _, id := range datasetIDs[start:end] {
			batch = append(batch, &GroupDataset{ID: id})
		}
		if err := g.send(ctx, method, url, batch); err != nil {
			return withRequestIDs(err, datasetIDs[start:end])
		}
	}
	return nil
}

// withRequestIDs names the ID behind each error in err. The API names a
// rejected user or dataset only by its position in the request, as the
// pointer "/data/<index>". ids are the IDs that request sent, in order.
func withRequestIDs(err error, ids []string) error {
	var de hnyclient.DetailedError
	if !errors.As(err, &de) {
		return err
	}
	de.Details = slices.Clone(de.Details)
	for i, d := range de.Details {
		index, ok := strings.CutPrefix(d.Field, "/data/")
		if !ok {
			continue
		}
		n, convErr := strconv.Atoi(index)
		if convErr != nil || n < 0 || n >= len(ids) {
			continue
		}
		// the position means nothing to the caller, so drop it
		de.Details[i].Field = ""
		de.Details[i].Description = ids[n] + ": " + d.Description
	}
	return de
}

// send makes a request whose success response is 204 No Content.
func (g *groups) send(ctx context.Context, method, url string, body any) error {
	r, err := g.client.Do(ctx, method, url, body)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusNoContent {
		return hnyclient.ErrorFromResponse(r)
	}
	return nil
}
