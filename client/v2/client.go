package v2

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/jsonapi"

	hnyclient "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/client/internal/limits"
)

const (
	DefaultAPIHost         = "https://api.honeycomb.io"
	DefaultAPIEndpointEnv  = "HONEYCOMB_API_ENDPOINT"
	DefaultAPIKeyIDEnv     = "HONEYCOMB_KEY_ID"
	DefaultAPIKeySecretEnv = "HONEYCOMB_KEY_SECRET"

	defaultUserAgent = "go-honeycomb"
)

type Config struct {
	APIKeyID     string
	APIKeySecret string
	BaseURL      string
	Debug        bool
	HTTPClient   *http.Client
	UserAgent    string
}

type Client struct {
	BaseURL   *url.URL
	Headers   http.Header
	UserAgent string

	http *retryablehttp.Client

	// authInfo is populated on first use by teamSlug. It caches the team slug
	// needed to construct v2 API paths so we don't do a blocking round-trip
	// during client construction.
	authMu   sync.Mutex
	authInfo *AuthMetadata

	// API handlers here
	APIKeys      APIKeys
	Environments Environments
}

// teamSlug returns the team slug for the configured API key, fetching the
// auth metadata from the API on first use and caching the result. Errors are
// not cached so that a transient failure does not poison later calls.
func (c *Client) teamSlug(ctx context.Context) (string, error) {
	c.authMu.Lock()
	defer c.authMu.Unlock()
	if c.authInfo == nil {
		info, err := c.AuthInfo(ctx)
		if err != nil {
			return "", err
		}
		c.authInfo = info
	}
	return c.authInfo.Team.Slug, nil
}

func NewClient() (*Client, error) {
	return NewClientWithConfig(nil)
}

func NewClientWithConfig(config *Config) (*Client, error) {
	if config == nil {
		config = &Config{}
	}
	if config.BaseURL == "" {
		host := os.Getenv(DefaultAPIEndpointEnv)
		if host == "" {
			config.BaseURL = DefaultAPIHost
		} else {
			config.BaseURL = host
		}
	}
	baseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid BaseURL: %w", err)
	}
	if config.UserAgent == "" {
		config.UserAgent = defaultUserAgent
	}

	if config.APIKeyID == "" && config.APIKeySecret == "" {
		config.APIKeyID = os.Getenv(DefaultAPIKeyIDEnv)
		config.APIKeySecret = os.Getenv(DefaultAPIKeySecretEnv)

		// if we still don't have the API key, we'll need to error out
		if config.APIKeyID == "" || config.APIKeySecret == "" {
			return nil, errors.New("missing API Key ID and Secret pair")
		}
	}
	token := config.APIKeyID + ":" + config.APIKeySecret

	client := &Client{
		UserAgent: config.UserAgent,
		BaseURL:   baseURL,
		Headers: http.Header{
			"Authorization": {"Bearer " + token},
			"Content-Type":  {jsonapi.MediaType},
			"User-Agent":    {config.UserAgent},
		},
	}
	client.http = &retryablehttp.Client{
		Backoff:      limits.RetryHTTPBackoff,
		CheckRetry:   limits.RetryHTTPCheck,
		ErrorHandler: retryablehttp.PassthroughErrorHandler,
		HTTPClient:   config.HTTPClient,
		RetryWaitMin: 200 * time.Millisecond,
		RetryWaitMax: time.Minute,
		RetryMax:     10,
	}

	if config.Debug {
		// if enabled we log all requests and responses to sterr
		client.http.Logger = log.New(os.Stderr, "", log.LstdFlags)
		client.http.ResponseLogHook = func(l retryablehttp.Logger, resp *http.Response) {
			l.Printf("[DEBUG] Request: %s %s", resp.Request.Method, resp.Request.URL.String())
			// TODO: Log request body
		}
	}

	// bind API handlers here
	client.APIKeys = &apiKeys{client: client}
	client.Environments = &environments{client: client}

	return client, nil
}

func (c *Client) Do(
	ctx context.Context,
	method,
	path string,
	body any,
) (*http.Response, error) {
	url, err := c.BaseURL.Parse(path)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest(
		ctx,
		method,
		url.String(),
		body,
	)
	if err != nil {
		return nil, err
	}

	return c.http.Do(req)
}

func (c *Client) AuthInfo(ctx context.Context) (*AuthMetadata, error) {
	r, err := c.Do(ctx, http.MethodGet, "/2/auth", nil)
	if err != nil {
		return nil, err
	}
	if r.StatusCode != http.StatusOK {
		return nil, hnyclient.ErrorFromResponse(r)
	}

	auth := new(AuthMetadata)
	if err := jsonapi.UnmarshalPayload(r.Body, auth); err != nil {
		return nil, err
	}
	return auth, err
}

func (c *Client) newRequest(
	ctx context.Context,
	method,
	url string,
	body any,
) (*retryablehttp.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		buf := bytes.NewBuffer(nil)
		if err := jsonapi.MarshalPayloadWithoutIncluded(buf, body); err != nil {
			return nil, err
		}
		bodyReader = buf
	}
	req, err := retryablehttp.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	for k, h := range c.Headers {
		req.Header[k] = append(req.Header[k], h...)
	}

	return req, err
}
