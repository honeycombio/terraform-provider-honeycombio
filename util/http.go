package util

import (
	"log"
	"net/http"
	"net/http/httputil"
)

// Is2xx checks whether the HTTP status code is successful.
func Is2xx(status int) bool {
	return status >= 200 && status < 300
}

// roundTripFunc implements http.RoundTripper
type roundTripFunc func(*http.Request) (*http.Response, error)

func (rf roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return rf(req)
}

// WrapWithLogging wraps the HTTP client to log the entire request and
// response.
func WrapWithLogging(c *http.Client) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			b, _ := httputil.DumpRequestOut(req, true)
			log.Printf("Sending request\n\n%s\n", b)

			res, err := c.Do(req)
			if err != nil {
				return res, err
			}

			b, _ = httputil.DumpResponse(res, true)
			log.Printf("Received response\n\n%s\n", b)

			return res, nil
		}),
	}
}
