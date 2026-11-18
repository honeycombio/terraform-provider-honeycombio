package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"

	hnyclient "github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestDetailedErrorDiagnostic_Detail(t *testing.T) {
	d := NewDetailedErrorDiagnostic("Updating X", &hnyclient.DetailedError{
		Status: 422,
		Details: []hnyclient.ErrorTypeDetail{
			{Code: "invalid", Description: "first"},
			{Code: "invalid", Description: "second"},
		},
	})

	assert.Equal(t, "invalid - first\ninvalid - second", d.Detail())
}
