package client

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hashicorp/jsonapi"
)

// DetailedError is an RFC7807 'Problem Detail' formatted error message.
type DetailedError struct {
	// The HTTP status code of the error.
	Status int `json:"status,omitempty"`
	// The error message
	Message string `json:"error,omitempty"`
	// ID is the unique ID of the HTTP request that caused this error.
	ID string `json:"request_id,omitempty"`
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

// IsNotFound returns true if the error is an HTTP 404
func (e *DetailedError) IsNotFound() bool {
	if e == nil {
		return false
	}
	return e.Status == http.StatusNotFound
}

// Error returns a pretty-printed representation of the error
func (e DetailedError) Error() string {
	response := ""
	if len(e.Details) > 0 {
		for index, details := range e.Details {
			response += details.String()

			// If we haven't reached the end of the list of error details,
			// add a newline separator between each error
			if index < len(e.Details)-1 {
				response += "\n"
			}
		}

		return response
	}

	return e.Message
}

func ErrorFromResponse(r *http.Response) error {
	if r == nil {
		return errors.New("invalid response")
	}

	requestID := r.Header.Get("Request-Id")

	switch r.Header.Get("Content-Type") {
	case jsonapi.MediaType:
		var detailedError DetailedError

		errPayload := new(jsonapi.ErrorsPayload)
		err := json.NewDecoder(r.Body).Decode(errPayload)
		if err != nil || len(errPayload.Errors) == 0 {
			return DetailedError{
				Status:  r.StatusCode,
				Message: r.Status,
				ID:      requestID,
			}
		}

		detailedError = DetailedError{
			ID:     requestID,
			Status: r.StatusCode,
			Title:  errPayload.Errors[0].Title,
		}

		if len(errPayload.Errors) == 1 {
			// format the detailed error a bit differently if there's only one error
			detailedError.Type = errPayload.Errors[0].Code
			detailedError.Details = []ErrorTypeDetail{
				{
					Description: errPayload.Errors[0].Title,
					Field:       parseJSONAPIErrorSource(errPayload.Errors[0].Source),
				},
			}
		} else {
			details := make([]ErrorTypeDetail, len(errPayload.Errors))
			for i, e := range errPayload.Errors {
				details[i] = ErrorTypeDetail{
					Description: e.Detail,
					Field:       parseJSONAPIErrorSource(e.Source),
				}
			}
			detailedError.Details = details
		}
		return detailedError
	default:
		var detailedError DetailedError
		if err := json.NewDecoder(r.Body).Decode(&detailedError); err != nil {
			// If we can't decode the error, return a generic error
			return DetailedError{
				ID:      requestID,
				Status:  r.StatusCode,
				Message: r.Status,
			}
		}

		// sanity check: ensure we have the status code set
		if detailedError.Status == 0 {
			detailedError.Status = r.StatusCode
		}

		// ensure we have the requestID set
		if detailedError.ID == "" {
			detailedError.ID = requestID
		}

		return detailedError
	}
}

// parseJSONAPIErrorSource returns a string representation of the
// source of a JSON:API Error Source
func parseJSONAPIErrorSource(e *jsonapi.ErrorSource) string {
	if e == nil {
		return ""
	}

	// the JSON:API specification states that only one of these
	// fields should be populated so we return the first that is not empty
	if e.Pointer != "" {
		return e.Pointer
	}
	if e.Parameter != "" { // not currently in use server-side
		return "parameter " + e.Parameter
	}
	if e.Header != "" {
		return e.Header + " header"
	}

	return ""
}
