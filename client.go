package honeycombiosdk

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

const (
	apiURL    = "https://api.honeycomb.io"
	userAgent = "terraform-provider-honeycombio"
)

// Client to interact with Honeycomb.
type Client struct {
	apiKey     string
	dataset    string
	httpClient *http.Client

	Markers Markers
}

// NewClient creates a new Honeycomb API client.
func NewClient(apiKey, dataset string) *Client {
	client := &Client{
		apiKey:     apiKey,
		dataset:    dataset,
		httpClient: &http.Client{},
	}
	client.Markers = &markers{client: client}

	return client
}

func (c *Client) newRequest(method, path string, v interface{}) (*http.Request, error) {
	var body io.Reader

	if v != nil {
		buf := new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(v)
		if err != nil {
			return nil, err
		}
		body = buf
	}

	req, err := http.NewRequest(method, apiURL+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-Honeycomb-Team", c.apiKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", userAgent)

	return req, nil
}

func is2xx(status int) bool {
	return status >= 200 && status < 300
}
