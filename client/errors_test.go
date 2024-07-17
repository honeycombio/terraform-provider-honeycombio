package client_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/jsonapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestClient_ParseDetailedError(t *testing.T) {
	t.Parallel()

	var de client.DetailedError
	ctx := context.Background()
	c := newTestClient(t)

	t.Run("Post with no body should fail with 400 unparseable", func(t *testing.T) {
		err := c.Do(ctx, "POST", "/1/boards/", nil, nil)
		require.Error(t, err)
		require.ErrorAs(t, err, &de)
		assert.Equal(t, http.StatusBadRequest, de.Status)
		assert.Equal(t, fmt.Sprintf("%s/problems/unparseable", c.EndpointURL()), de.Type)
		assert.Equal(t, "The request body could not be parsed.", de.Title)
		assert.Equal(t, "could not parse request body", de.Message)
	})

	t.Run("Get into non-existent dataset should fail with 404 'Dataset not found'", func(t *testing.T) {
		_, err := c.Markers.Get(ctx, "non-existent-dataset", "abcd1234")
		require.Error(t, err)
		require.ErrorAs(t, err, &de)
		assert.Equal(t, http.StatusNotFound, de.Status)
		assert.Equal(t, fmt.Sprintf("%s/problems/not-found", c.EndpointURL()), de.Type)
		assert.Equal(t, "The requested resource cannot be found.", de.Title)
		assert.Equal(t, "Dataset not found", de.Message)
	})

	t.Run("Creating a dataset without a name should return a validation error", func(t *testing.T) {
		createDatasetRequest := &client.Dataset{}
		_, err := c.Datasets.Create(ctx, createDatasetRequest)
		require.Error(t, err)
		require.ErrorAs(t, err, &de)
		assert.Equal(t, http.StatusUnprocessableEntity, de.Status)
		assert.Equal(t, fmt.Sprintf("%s/problems/validation-failed", c.EndpointURL()), de.Type)
		assert.Equal(t, "The provided input is invalid.", de.Title)
		assert.Equal(t, "The provided input is invalid.", de.Message)
		assert.Len(t, de.Details, 1)
		assert.Equal(t, "missing", de.Details[0].Code)
		assert.Equal(t, "name", de.Details[0].Field)
		assert.Equal(t, "cannot be blank", de.Details[0].Description)
		assert.Equal(t, "missing name - cannot be blank", de.Error())
	})
}

func TestErrors_DetailedError_Error(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          client.DetailedError
		expectedOutput string
	}{
		{
			name: "multiple details get separated by newline",
			input: client.DetailedError{
				Message: "test message",
				Details: []client.ErrorTypeDetail{
					{
						Code:        "test code1",
						Field:       "test_field1",
						Description: "test description1",
					},
					{
						Code:        "test code2",
						Field:       "test_field2",
						Description: "test description2",
					},
				},
			},
			expectedOutput: "test code1 test_field1 - test description1\ntest code2 test_field2 - test description2",
		},
		{
			name: "empty details returns message",
			input: client.DetailedError{
				Message: "test message",
				Details: []client.ErrorTypeDetail{},
			},
			expectedOutput: "test message",
		},
		{
			name: "one item in details has no newlines",
			input: client.DetailedError{
				Message: "test message",
				Details: []client.ErrorTypeDetail{
					{
						Code:        "test code",
						Field:       "test_field",
						Description: "test description",
					},
				},
			},
			expectedOutput: "test code test_field - test description",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualOutput := testCase.input.Error()
			assert.Equal(t, testCase.expectedOutput, actualOutput)
		})
	}
}

func TestErrors_ErrorTypeDetail_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          client.ErrorTypeDetail
		expectedOutput string
	}{
		{
			name: "happy path: Code, Field, and Description present",
			input: client.ErrorTypeDetail{
				Code:        "test code",
				Field:       "test_field",
				Description: "test description",
			},
			expectedOutput: "test code test_field - test description",
		},
		{
			name:           "all fields blank returns empty string",
			input:          client.ErrorTypeDetail{},
			expectedOutput: "",
		},
		{
			name: "empty Code",
			input: client.ErrorTypeDetail{
				Field:       "test_field",
				Description: "test description",
			},
			expectedOutput: "test_field - test description",
		},
		{
			name: "empty Code and Field",
			input: client.ErrorTypeDetail{
				Description: "test description",
			},
			expectedOutput: "test description",
		},
		{
			name: "empty Code and Description",
			input: client.ErrorTypeDetail{
				Field: "test_field",
			},
			expectedOutput: "test_field",
		},
		{
			name: "empty Field",
			input: client.ErrorTypeDetail{
				Code:        "test code",
				Description: "test description",
			},
			expectedOutput: "test code test description",
		},
		{
			name: "empty Field and Description",
			input: client.ErrorTypeDetail{
				Code: "test code",
			},
			expectedOutput: "test code",
		},
		{
			name: "empty Description",
			input: client.ErrorTypeDetail{
				Code:  "test code",
				Field: "test_field",
			},
			expectedOutput: "test code test_field",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualOutput := testCase.input.String()
			assert.Equal(t, testCase.expectedOutput, actualOutput)
		})
	}
}

func TestErrors_JSONAPI(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		code           int
		body           jsonapi.ErrorsPayload
		expectedOutput client.DetailedError
	}{
		{
			name: "single error",
			code: http.StatusConflict,
			body: jsonapi.ErrorsPayload{
				Errors: []*jsonapi.ErrorObject{
					{
						Status: "409",
						Title:  "Conflict",
						Detail: "The resource already exists.",
						Code:   "/errors/conflict",
					},
				},
			},
			expectedOutput: client.DetailedError{
				Status:  409,
				Type:    "/errors/conflict",
				Message: "The resource already exists.",
				Title:   "Conflict",
			},
		},
		{
			name: "multi error",
			code: http.StatusUnprocessableEntity,
			body: jsonapi.ErrorsPayload{
				Errors: []*jsonapi.ErrorObject{
					{
						Status: "422",
						Code:   "/errors/validation-failed",
						Title:  "The provided input is invalid.",
					},
					{
						Status: "422",
						Code:   "/errors/validation-failed",
						Title:  "The provided input is invalid.",
					},
				},
			},
			expectedOutput: client.DetailedError{
				Status: 422,
				Title:  "The provided input is invalid.",
				Details: []client.ErrorTypeDetail{
					{
						Code: "/errors/validation-failed",
					},
					{
						Code: "/errors/validation-failed",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resp := &httptest.ResponseRecorder{
				Code: testCase.code,
				HeaderMap: http.Header{
					"Content-Type": []string{jsonapi.MediaType},
				},
			}

			buf := bytes.NewBuffer(nil)
			jsonapi.MarshalErrors(buf, testCase.body.Errors)
			resp.Body = bytes.NewBuffer(buf.Bytes())

			actualOutput := client.ErrorFromResponse(resp.Result())
			assert.Equal(t, testCase.expectedOutput, actualOutput)
		})
	}
}
