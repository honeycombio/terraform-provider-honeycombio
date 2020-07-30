package honeycombio

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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

	Markers  Markers
	Triggers Triggers
}

// NewClient creates a new Honeycomb API client.
func NewClient(apiKey, dataset string) *Client {
	client := &Client{
		apiKey:     apiKey,
		dataset:    dataset,
		httpClient: &http.Client{},
	}
	client.Markers = &markers{client: client}
	client.Triggers = &triggers{client: client}

	return client
}

// ErrNotFound means that the requested item could not be found.
var ErrNotFound = errors.New("request failed with status code 404")

// newRequest prepares a request to the Honeycomb API with the default Honeycomb
// headers and a JSON body, if v is set.
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

// do a request and parse the response in v, if v is not nil. Returns an error
// if the request failed or if the response contained a non-2xx status code.
// ErrNotFound is returned on a 404 response.
func (c *Client) do(req *http.Request, v interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !is2xx(resp.StatusCode) {
		if resp.StatusCode == 404 {
			return ErrNotFound
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("request failed with status code %d", resp.StatusCode)
		}
		return fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, body)
	}

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
	}
	return err
}

func is2xx(status int) bool {
	return status >= 200 && status < 300
}
