package helper

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

// DetailedErrorDiagnostic is a Diagnostic which nicely wraps a client.DetailedError
type DetailedErrorDiagnostic struct {
	summary string
	e       *client.DetailedError
}

// compile-time check that DetailedErrorDiagnostic implements diag.Diagnostic
var _ diag.Diagnostic = DetailedErrorDiagnostic{}

// NewDetailedErrorDiagnostic creates a new DetailedErrorDiagnostic
// taking a context-specific summary of the action and a DetailedError.
func NewDetailedErrorDiagnostic(summary string, e *client.DetailedError) DetailedErrorDiagnostic {
	return DetailedErrorDiagnostic{
		e:       e,
		summary: summary,
	}
}

// AddDiagnosticOnError is a helper function which will take an error
// and a context-specific summary of the action.
//
// If err is nil, this function will return false.
// If err is non-nil, this function will return ture and take one the following actions:
//   - if it is a client.DetailedError, a DetailedErrorDiagnostic will be added to diag.
//   - otherwise a generic error diagnostic will be added to diag.
func AddDiagnosticOnError(diag *diag.Diagnostics, summary string, err error) bool {
	if err == nil {
		return false
	}

	var detailedErr *client.DetailedError
	if errors.As(err, &detailedErr) {
		diag.Append(DetailedErrorDiagnostic{
			summary: "Error " + summary,
			e:       detailedErr,
		})
	} else {
		diag.AddError("Error "+summary, err.Error())
	}
	return true
}

func (d DetailedErrorDiagnostic) Detail() string {
	if len(d.e.Details) > 0 {
		var response string
		for i, dt := range d.e.Details {
			response += dt.Code + " - " + dt.Description

			if i > len(d.e.Details)-1 {
				response += "\n "
			}
		}
		return response
	}

	return d.e.Message
}

func (d DetailedErrorDiagnostic) Summary() string {
	return fmt.Sprintf("%s (HTTP %d) - %s", d.summary, d.e.Status, d.e.Title)
}

func (d DetailedErrorDiagnostic) Severity() diag.Severity {
	return diag.SeverityError
}

func (d DetailedErrorDiagnostic) Equal(other diag.Diagnostic) bool {
	ed, ok := other.(DetailedErrorDiagnostic)
	if !ok {
		return false
	}

	return d.Summary() == ed.Summary() && reflect.DeepEqual(d, ed)
}
