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
	if len(e.Details) > 0 {
		var response string

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

	switch r.Header.Get("Content-Type") {
	case jsonapi.MediaType:
		var detailedError DetailedError

		errPayload := new(jsonapi.ErrorsPayload)
		err := json.NewDecoder(r.Body).Decode(errPayload)
		if err != nil || len(errPayload.Errors) == 0 {
			return DetailedError{
				Status:  r.StatusCode,
				Message: r.Status,
			}
		}

		detailedError = DetailedError{
			Status: r.StatusCode,
			Title:  errPayload.Errors[0].Title,
		}
		if len(errPayload.Errors) == 1 {
			// If there's only one error we don't need to build up details
			detailedError.Message = errPayload.Errors[0].Detail
			detailedError.Type = errPayload.Errors[0].Code
		} else {
			details := make([]ErrorTypeDetail, len(errPayload.Errors))
			for i, e := range errPayload.Errors {
				details[i] = ErrorTypeDetail{
					Code:        e.Code,
					Description: e.Detail,
					// TODO: field when we have it via pointer
				}
			}
		}
		return detailedError
	default:
		var detailedError DetailedError
		if err := json.NewDecoder(r.Body).Decode(&detailedError); err != nil {
			// If we can't decode the error, return a generic error
			return DetailedError{
				Status:  r.StatusCode,
				Message: r.Status,
			}
		}

		// sanity check: ensure we have the status code set
		if detailedError.Status == 0 {
			detailedError.Status = r.StatusCode
		}
		return detailedError
	}
}
