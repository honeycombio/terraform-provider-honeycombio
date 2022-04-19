package client

import (
	"context"
	"fmt"
	"time"
)

// QueryResults describes all the query result-related methods that the
// Honeycomb API supports.
//
// API docs: https://docs.honeycomb.io/api/query-results/
type QueryResults interface {
	// Get the query results by ID. Returns ErrNotFound if there is no
	// query result with the given ID.
	Get(ctx context.Context, dataset string, q *QueryResult) error

	// Create a new query result with a given query specification.
	Create(ctx context.Context, dataset string, data *QueryResultRequest) (*QueryResult, error)
}

// queryResults implements QueryResults.
type queryResults struct {
	client *Client
}

// Compile-time proof of interface implementation by type queryResults.
var _ QueryResults = (*queryResults)(nil)

// QueryResult represents a Honeycomb query result.
//
// API docs: https://docs.honeycomb.io/api/query-results/#get-example-response
type QueryResult struct {
	// ID of a query result is only set when the query is returned from the
	// Query Result API. This value should not be set when creating queries results.
	ID string `json:"id"`

	// True once the query has completed and the results are populated.
	Complete bool `json:"complete"`

	// The resulting data of the query
	Data QueryResultData `json:"data,omitempty"`

	// Permalinks to the query results
	Links QueryResultLinks `json:"links,omitempty"`
}

type QueryResultRequest struct {
	ID string `json:"query_id"`
}

type QueryResultData struct {
	Series []struct {
		Time time.Time              `json:"time"`
		Data map[string]interface{} `json:"data"`
	} `json:"series"`
	Results []struct {
		Data map[string]interface{} `json:"data"`
	} `json:"results"`
}

const QueryResultPollInterval time.Duration = 200 * time.Millisecond

type QueryResultLinks struct {
	Url      string `json:"query_url"`
	GraphUrl string `json:"graph_image_url"`
}

func (s *queryResults) Get(ctx context.Context, dataset string, q *QueryResult) error {
	var err error
	resultUri := fmt.Sprintf("/1/query_results/%s/%s", urlEncodeDataset(dataset), q.ID)

	ticker := time.NewTicker(QueryResultPollInterval)
	for ; ; <-ticker.C {
		// poll until complete or errored
		if err = s.client.performRequest(ctx, "GET", resultUri, nil, &q); err != nil {
			return err
		}
		if q.Complete {
			break
		}
	}
	return nil
}

func (s *queryResults) Create(ctx context.Context, dataset string, data *QueryResultRequest) (*QueryResult, error) {
	var q QueryResult
	err := s.client.performRequest(ctx, "POST", fmt.Sprintf("/1/query_results/%s", urlEncodeDataset(dataset)), data, &q)
	return &q, err
}
