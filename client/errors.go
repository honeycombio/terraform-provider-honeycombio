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
	Details []ErrorTypeDetail `json:"type_detail,omitempty"`
}

type ErrorTypeDetail struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Field       string `json:"field"`
}

func (td ErrorTypeDetail) String() string {
	response := ""
	if td.Code != "" {
		response += td.Code
		// If any of the fields that could come next are present, add a space separator
		if td.Field != "" || td.Description != "" {
			response += " "
		}
	}

	if td.Field != "" {
		response += td.Field
		// If any of the fields that could come next are present, add a dash separator
		if td.Description != "" {
			response += " - "
		}
	}

	if td.Description != "" {
		response += td.Description
	}

	return response
}

// Error returns a pretty-printed representation of the error
func (e DetailedError) Error() string {
	if len(e.Details) > 0 {
		var response string

		for index, details := range e.Details {
			response += details.String()

			// If we haven't reached the end of the list of error details, add a newline separator between each error
			if index < len(e.Details)-1 {
				response += "\n"
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
		return DetailedError{
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
