package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// DetailedError is an RFC7807 'Problem Detail' formatted error message.
type DetailedError struct {
	// The HTTP status code of the error.
	Status int `json:"status,omitempty"`
	// The error message
	Message string `json:"error,omitempty"`
	// Type is a URI used to uniquely identify the type of error.
	Type string `json:"type,omitempty"`
	// Title is a human-readable summary that explains the type of the problem.
	Title string `json:"title,omitempty"`
	// Details is an array of structured objects that give details about the error.
	Details []struct {
		Code        string `json:"code"`
		Description string `json:"description"`
		Field       string `json:"field"`
	} `json:"type_detail,omitempty"`
}

// Error returns a pretty-printed representation of the error
func (e DetailedError) Error() string {
	if len(e.Details) > 0 {
		var response string
		for i, d := range e.Details {
			response += d.Code + " - " + d.Description
			if i > len(e.Details)-1 {
				response += ", "
			}
		}
		return response
	}

	return e.Message
}

// IsNotFound returns true if the error is an HTTP 404
func (e *DetailedError) IsNotFound() bool {
	if e == nil {
		return false
	}
	return e.Status == http.StatusNotFound
}

func errorFromResponse(resp *http.Response) error {
	e, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body: %w", err)
	}

	var detailedErr DetailedError
	err = json.Unmarshal(e, &detailedErr)
	if err != nil {
		// we failed to parse the body as a DetailedError, so build one from what we know
		return &DetailedError{
			Status:  resp.StatusCode,
			Message: resp.Status,
		}
	}

	// quick sanity check to make sure we got a StatusCode
	if detailedErr.Status == 0 {
		detailedErr.Status = resp.StatusCode
	}

	return detailedErr
}
