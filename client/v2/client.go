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
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/jsonapi"

	hnyclient "github.com/honeycombio/terraform-provider-honeycombio/client"
)

const (
	DefaultAPIHost         = "https://api.honeycomb.io"
	DefaultAPIEndpointEnv  = "HONEYCOMB_API_ENDPOINT"
	DefaultAPIKeyIDEnv     = "HONEYCOMB_KEY_ID"
	DefaultAPIKeySecretEnv = "HONEYCOMB_KEY_SECRET"

	defaultUserAgent = "go-honeycomb"
)

type Config struct {
	APIKeyID           string
	APIKeySecret       string
	BaseURL            string
	Debug              bool
	HTTPClient         *http.Client
	UserAgent          string
	skipInitialization bool
}

type Client struct {
	BaseURL   *url.URL
	Headers   http.Header
	UserAgent string

	http *retryablehttp.Client

	// API handlers here
	APIKeys      APIKeys
	Environments Environments
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
		Backoff:      client.retryHTTPBackoff,
		CheckRetry:   client.retryHTTPCheck,
		ErrorHandler: retryablehttp.PassthroughErrorHandler,
		HTTPClient:   config.HTTPClient,
		RetryWaitMin: 200 * time.Millisecond,
		RetryWaitMax: 10 * time.Second,
		RetryMax:     15,
	}

	if config.Debug {
		// if enabled we log all requests and responses to sterr
		client.http.Logger = log.New(os.Stderr, "", log.LstdFlags)
		client.http.ResponseLogHook = func(l retryablehttp.Logger, resp *http.Response) {
			l.Printf("[DEBUG] Request: %s %s", resp.Request.Method, resp.Request.URL.String())
			// TODO: Log request body
		}
	}

	// early out if we're just creating the client for testing
	if config.skipInitialization {
		return client, nil
	}

	var authinfo *AuthMetadata
	authinfo, err = client.AuthInfo(context.Background())
	if err != nil {
		return nil, err
	}

	// bind API handlers here
	client.APIKeys = &apiKeys{client: client, authinfo: authinfo}
	client.Environments = &environments{client: client, authinfo: authinfo}

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

// retryHTTPCheck is a retryablehttp.CheckRetry function that will
// retry on a 420 or any 5xx status code
func (c *Client) retryHTTPCheck(
	ctx context.Context,
	r *http.Response,
	err error,
) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	if err != nil {
		return true, err
	}
	if r != nil {
		if r.StatusCode == http.StatusTooManyRequests || r.StatusCode >= 500 {
			return true, nil
		}
	}
	return false, nil
}

// retryHTTPBackoff is a retryablehttp.Backoff function that will
// use a linear backoff for all status codes except 429, which will
// attempt to use the rate limit headers to determine the backoff time
func (c *Client) retryHTTPBackoff(
	min, max time.Duration,
	attemptNum int,
	r *http.Response,
) time.Duration {
	if r != nil && r.StatusCode == http.StatusTooManyRequests {
		return rateLimitBackoff(min, max, r)
	}

	// if we've not been rate limited, use a linear backoff
	// but increase the minimum and maximum backoff times
	// and hand it off to retryablehttp.LinearJitterBackoff
	min = 500 * time.Millisecond
	max = 950 * time.Millisecond
	return retryablehttp.LinearJitterBackoff(min, max, attemptNum, r)
}
