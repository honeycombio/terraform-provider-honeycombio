// Package client provides a client to interact with the Honeycomb API.
//
// Documentation of the API can be found here: https://docs.honeycomb.io/api/
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	cleanhttp "github.com/hashicorp/go-cleanhttp"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

const (
	DefaultAPIHost        = "https://api.honeycomb.io"
	DefaultAPIEndpointEnv = "HONEYCOMB_API_ENDPOINT"
	DefaultAPIKeyEnv      = "HONEYCOMB_API_KEY"
	// Deprecated: use DefaultAPIKeyEnv instead. To be removed in v1.0
	LegacyAPIKeyEnv  = "HONEYCOMBIO_APIKEY"
	defaultUserAgent = "go-honeycombio"
)

// Config holds all configuration options for the client.
type Config struct {
	// Required - the API key to use when sending request to Honeycomb.
	APIKey string
	// URL of the Honeycomb API, defaults to "https://api.honeycomb.io".
	APIUrl string
	// With debug enabled the client will log all requests and responses.
	Debug bool
	// Optionally override the HTTP client with a custom client.
	HTTPClient *http.Client
	// Optionally set the user agent to send with all requests, defaults to "go-honeycombio".
	UserAgent string
}

// Client to interact with Honeycomb.
type Client struct {
	apiKey     string
	apiURL     *url.URL
	headers    http.Header
	httpClient *retryablehttp.Client

	Auth               Auth
	Boards             Boards
	Columns            Columns
	Datasets           Datasets
	DatasetDefinitions DatasetDefinitions
	DerivedColumns     DerivedColumns
	Markers            Markers
	MarkerSettings     MarkerSettings
	Queries            Queries
	QueryAnnotations   QueryAnnotations
	QueryResults       QueryResults
	Triggers           Triggers
	SLOs               SLOs
	BurnAlerts         BurnAlerts
	Recipients         Recipients
}

// DefaultConfig returns a Config initilized with default values.
func DefaultConfig() *Config {
	c := &Config{
		APIKey:     os.Getenv(DefaultAPIKeyEnv),
		APIUrl:     os.Getenv(DefaultAPIEndpointEnv),
		Debug:      false,
		HTTPClient: cleanhttp.DefaultPooledClient(),
		UserAgent:  defaultUserAgent,
	}

	// if API Key is still unset, try using the legacy environment variable
	if c.APIKey == "" {
		c.APIKey = os.Getenv(LegacyAPIKeyEnv)
	}

	if c.APIUrl == "" {
		c.APIUrl = DefaultAPIHost
	}

	return c
}

// NewClient creates a new Honeycomb API client with default settings.
func NewClient() (*Client, error) {
	return NewClientWithConfig(DefaultConfig())
}

// NewClientWithConfig creates a new Honeycomb API client using the provided
// Config.
func NewClientWithConfig(config *Config) (*Client, error) {
	cfg := DefaultConfig()

	if config.APIKey != "" {
		cfg.APIKey = config.APIKey
	}
	if config.APIUrl != "" {
		cfg.APIUrl = config.APIUrl
	}
	if config.UserAgent != "" {
		cfg.UserAgent = config.UserAgent
	}
	if config.HTTPClient != nil {
		cfg.HTTPClient = config.HTTPClient
	}

	if cfg.APIKey == "" {
		return nil, errors.New("APIKey must be configured")
	}
	apiURL, err := url.Parse(cfg.APIUrl)
	if err != nil {
		return nil, fmt.Errorf("could not parse APIUrl: %w", err)
	}

	client := &Client{
		apiKey:  cfg.APIKey,
		apiURL:  apiURL,
		headers: make(http.Header),
	}

	client.httpClient = &retryablehttp.Client{
		Backoff:      retryablehttp.DefaultBackoff,
		CheckRetry:   client.retryHTTPCheck,
		ErrorHandler: retryablehttp.PassthroughErrorHandler,
		HTTPClient:   cfg.HTTPClient,
		RetryWaitMin: 200 * time.Millisecond,
		RetryWaitMax: 10 * time.Second,
		RetryMax:     15,
	}

	if config.Debug {
		// if enabled we log all requests and responses to sterr
		client.httpClient.Logger = log.New(os.Stderr, "", log.LstdFlags)
		client.httpClient.ResponseLogHook = func(l retryablehttp.Logger, resp *http.Response) {
			l.Printf("[DEBUG] Response: %s %s", resp.Request.Method, resp.Request.URL.String())
		}
	}

	client.headers.Add("Content-Type", "application/json")
	client.headers.Add("User-Agent", cfg.UserAgent)
	client.headers.Add("X-Honeycomb-Team", cfg.APIKey)

	client.Auth = &auth{client: client}
	client.Boards = &boards{client: client}
	client.Columns = &columns{client: client}
	client.Datasets = &datasets{client: client}
	client.DatasetDefinitions = &datasetDefinitions{client: client}
	client.DerivedColumns = &derivedColumns{client: client}
	client.Markers = &markers{client: client}
	client.MarkerSettings = &markerSettings{client: client}
	client.Queries = &queries{client: client}
	client.QueryAnnotations = &queryAnnotations{client: client}
	client.QueryResults = &queryResults{client: client}
	client.Triggers = &triggers{client: client}
	client.SLOs = &slos{client: client}
	client.BurnAlerts = &burnalerts{client: client}
	client.Recipients = &recipients{client: client}

	return client, nil
}

// EndpointURL returns the Client's configured API endpoint URL
func (c *Client) EndpointURL() *url.URL {
	return c.apiURL
}

// IsClassic returns true if the client is configured with a Classic API Key
//
// If there is an error fetching the auth metadata, this will return false.
func (c *Client) IsClassic(ctx context.Context) bool {
	metadata, err := c.Auth.List(ctx)
	if err != nil {
		return false
	}
	return metadata.Environment.Slug == ""
}

// Do makes a request to the configured Honeycomb API endpoint
// and, if requestBody is not nil, sends along the JSON.
//
// The response is parsed in responseBody, if responseBody is not nil.
//
// Attempts to return a DetailedError if the response status code is not 2xx,
// but can return a generic error.
func (c *Client) Do(ctx context.Context, method, path string, requestBody, responseBody interface{}) error {
	var body io.Reader

	if requestBody != nil {
		buf := new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(requestBody)
		if err != nil {
			return err
		}
		body = buf
	}

	requestURL, err := c.apiURL.Parse(path)
	if err != nil {
		return err
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, method, requestURL.String(), body)
	if err != nil {
		return err
	}
	req.Header = c.headers

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		return ErrorFromResponse(resp)
	}
	if responseBody != nil {
		err = json.NewDecoder(resp.Body).Decode(responseBody)
	}

	return err
}

// retryHTTPCheck provides a callback for Client.CheckRetry which
// will retry both rate limit (429) and server gateway (502, 504) errors.
func (c *Client) retryHTTPCheck(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	if err != nil {
		return true, err
	}

	if resp != nil {
		if resp.StatusCode == http.StatusTooManyRequests ||
			resp.StatusCode == http.StatusBadGateway ||
			resp.StatusCode == http.StatusGatewayTimeout {
			return true, nil
		}
	}

	return false, nil
}

// urlEncodeDataset sanitizes the dataset name for when it is used as part of
// the URL.
func urlEncodeDataset(dataset string) string {
	return strings.Replace(dataset, "/", "-", -1)
}
