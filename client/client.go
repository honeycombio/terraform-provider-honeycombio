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
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/honeycombio/terraform-provider-honeycombio/client/internal/httputil"
)

// Config holds all configuration options for the client.
type Config struct {
	// Required - the API key to use when sending request to Honeycomb.
	APIKey string
	// URL of the Honeycomb API, defaults to "https://api.honeycomb.io".
	APIUrl string
	// User agent to send with all requests, defaults to "go-honeycombio".
	UserAgent string
	// With debug enabled the client will log all requests and responses.
	Debug bool
}

func defaultConfig() *Config {
	return &Config{
		APIKey:    "",
		APIUrl:    "https://api.honeycomb.io",
		UserAgent: "go-honeycombio",
		Debug:     false,
	}
}

// Merge the given config by copying all non-blank values.
func (c *Config) merge(other *Config) {
	if other.APIKey != "" {
		c.APIKey = other.APIKey
	}
	if other.APIUrl != "" {
		c.APIUrl = other.APIUrl
	}
	if other.UserAgent != "" {
		c.UserAgent = other.UserAgent
	}
	if c.Debug || other.Debug {
		c.Debug = true
	}
}

// Client to interact with Honeycomb.
type Client struct {
	apiKey     string
	apiURL     *url.URL
	userAgent  string
	httpClient *http.Client

	Boards           Boards
	Columns          Columns
	Datasets         Datasets
	DerivedColumns   DerivedColumns
	Markers          Markers
	Queries          Queries
	QueryAnnotations QueryAnnotations
	QueryResults     QueryResults
	Triggers         Triggers
}

// NewClient creates a new Honeycomb API client.
func NewClient(config *Config) (*Client, error) {
	cfg := defaultConfig()
	cfg.merge(config)

	if cfg.APIKey == "" {
		return nil, errors.New("APIKey must be configured")
	}
	apiURL, err := url.Parse(cfg.APIUrl)
	if err != nil {
		return nil, fmt.Errorf("could not parse APIUrl: %w", err)
	}

	httpClient := &http.Client{}
	if cfg.Debug {
		httpClient = httputil.WrapWithLogging(httpClient)
	}

	client := &Client{
		apiKey:     cfg.APIKey,
		apiURL:     apiURL,
		userAgent:  cfg.UserAgent,
		httpClient: httpClient,
	}
	client.Boards = &boards{client: client}
	client.Columns = &columns{client: client}
	client.Datasets = &datasets{client: client}
	client.DerivedColumns = &derivedColumns{client: client}
	client.Markers = &markers{client: client}
	client.Queries = &queries{client: client}
	client.QueryAnnotations = &queryAnnotations{client: client}
	client.QueryResults = &queryResults{client: client}
	client.Triggers = &triggers{client: client}

	return client, nil
}

// ErrNotFound is returned when the requested item could not be found.
var ErrNotFound = errors.New("404 Not Found")

// performRequest against the Honeycomb API with the necessary headers and, if
// requestBody is not nil, a JSON body. The response is parsed in responseBody,
// if responseBody is not nil.
// Returns an error if the request failed, if the response contained a non-2xx
// status code or if parsing the response in responseBody failed. ErrNotFound
// is returned on a 404 response.
func (c *Client) performRequest(ctx context.Context, method, path string, requestBody, responseBody interface{}) error {
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

	req, err := http.NewRequestWithContext(ctx, method, requestURL.String(), body)
	if err != nil {
		return err
	}

	req.Header.Add("X-Honeycomb-Team", c.apiKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !httputil.Is2xx(resp.StatusCode) {
		if resp.StatusCode == 404 {
			return ErrNotFound
		}
		return errorFromResponse(resp)
	}

	if responseBody != nil {
		err = json.NewDecoder(resp.Body).Decode(responseBody)
	}
	return err
}

func errorFromResponse(resp *http.Response) error {
	errorMsg := attemptToExtractHoneycombioError(resp.Body)
	if errorMsg == "" {
		return fmt.Errorf("%s", resp.Status)
	}
	return fmt.Errorf("%s: %s", resp.Status, errorMsg)
}

type honeycombioError struct {
	ErrorMessage string `json:"error"`
}

func attemptToExtractHoneycombioError(bodyReader io.Reader) string {
	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return ""
	}

	var honeycombioErr honeycombioError

	err = json.Unmarshal(body, &honeycombioErr)
	if err != nil || honeycombioErr.ErrorMessage == "" {
		return string(body)
	}

	return honeycombioErr.ErrorMessage
}

// urlEncodeDataset sanitizes the dataset name for when it is used as part of
// the URL.
func urlEncodeDataset(dataset string) string {
	return strings.Replace(dataset, "/", "-", -1)
}
