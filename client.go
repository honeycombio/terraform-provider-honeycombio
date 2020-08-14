package honeycombio

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

	"github.com/kvrhdn/go-honeycombio/util"
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

	Boards   Boards
	Markers  Markers
	Triggers Triggers
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
		httpClient = util.WrapWithLogging(httpClient)
	}

	client := &Client{
		apiKey:     cfg.APIKey,
		apiURL:     apiURL,
		userAgent:  cfg.UserAgent,
		httpClient: httpClient,
	}
	client.Boards = &boards{client: client}
	client.Markers = &markers{client: client}
	client.Triggers = &triggers{client: client}

	return client, nil
}

// ErrNotFound means that the requested item could not be found.
var ErrNotFound = errors.New("404 Not Found")

// newRequest prepares a request to the Honeycomb API with the default Honeycomb
// headers and a JSON body, if v is set.
func (c *Client) newRequest(ctx context.Context, method, path string, v interface{}) (*http.Request, error) {
	var body io.Reader

	if v != nil {
		buf := new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(v)
		if err != nil {
			return nil, err
		}
		body = buf
	}

	requestURL, err := c.apiURL.Parse(path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-Honeycomb-Team", c.apiKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", c.userAgent)

	return req, nil
}

// do a request and parse the response in v, if v is not nil. Returns an error
// if the request failed or if the response contained a non-2xx status code.
// ErrNotFound is returned on a 404 response.
func (c *Client) do(req *http.Request, v interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !util.Is2xx(resp.StatusCode) {
		if resp.StatusCode == 404 {
			return ErrNotFound
		}
		return errorFromResponse(resp)
	}

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
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
