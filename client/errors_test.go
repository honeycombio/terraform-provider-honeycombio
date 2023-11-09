package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ParseDetailedError(t *testing.T) {
	var de DetailedError
	ctx := context.Background()
	c := newTestClient(t)

	t.Run("Post with no body should fail with 400 unparseable", func(t *testing.T) {
		err := c.performRequest(ctx, "POST", "/1/boards/", nil, nil)
		require.Error(t, err)
		assert.ErrorAs(t, err, &de)
		assert.Equal(t, de.Status, http.StatusBadRequest)
		assert.Equal(t, de.Type, "https://api.honeycomb.io/problems/unparseable")
		assert.Equal(t, de.Title, "The request body could not be parsed.")
		assert.Equal(t, de.Message, "could not parse request body")
	})

	t.Run("Get into non-existent dataset should fail with 404 'Dataset not found'", func(t *testing.T) {
		_, err := c.Markers.Get(ctx, "non-existent-dataset", "abcd1234")
		require.Error(t, err)
		assert.ErrorAs(t, err, &de)
		assert.Equal(t, de.Status, http.StatusNotFound)
		assert.Equal(t, de.Type, "https://api.honeycomb.io/problems/not-found")
		assert.Equal(t, de.Title, "The requested resource cannot be found.")
		assert.Equal(t, de.Message, "Dataset not found")
	})

	t.Run("Creating a dataset without a name should return a validation error", func(t *testing.T) {
		createDatasetRequest := &Dataset{}
		_, err := c.Datasets.Create(ctx, createDatasetRequest)
		require.Error(t, err)
		assert.ErrorAs(t, err, &de)
		assert.Equal(t, http.StatusUnprocessableEntity, de.Status)
		assert.Equal(t, "https://api.honeycomb.io/problems/validation-failed", de.Type)
		assert.Equal(t, "The provided input is invalid.", de.Title)
		assert.Equal(t, "The provided input is invalid.", de.Message)
		assert.Equal(t, 1, len(de.Details))
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
		input          DetailedError
		expectedOutput string
	}{
		{
			name: "multiple details get separated by newline",
			input: DetailedError{
				Message: "test message",
				Details: []ErrorTypeDetail{
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
			input: DetailedError{
				Message: "test message",
				Details: []ErrorTypeDetail{},
			},
			expectedOutput: "test message",
		},
		{
			name: "one item in details has no newlines",
			input: DetailedError{
				Message: "test message",
				Details: []ErrorTypeDetail{
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
	testCases := []struct {
		name           string
		input          ErrorTypeDetail
		expectedOutput string
	}{
		{
			name: "happy path: Code, Field, and Description present",
			input: ErrorTypeDetail{
				Code:        "test code",
				Field:       "test_field",
				Description: "test description",
			},
			expectedOutput: "test code test_field - test description",
		},
		{
			name:           "all fields blank returns empty string",
			input:          ErrorTypeDetail{},
			expectedOutput: "",
		},
		{
			name: "empty Code",
			input: ErrorTypeDetail{
				Field:       "test_field",
				Description: "test description",
			},
			expectedOutput: "test_field - test description",
		},
		{
			name: "empty Code and Field",
			input: ErrorTypeDetail{
				Description: "test description",
			},
			expectedOutput: "test description",
		},
		{
			name: "empty Code and Description",
			input: ErrorTypeDetail{
				Field: "test_field",
			},
			expectedOutput: "test_field",
		},
		{
			name: "empty Field",
			input: ErrorTypeDetail{
				Code:        "test code",
				Description: "test description",
			},
			expectedOutput: "test code test description",
		},
		{
			name: "empty Field and Description",
			input: ErrorTypeDetail{
				Code: "test code",
			},
			expectedOutput: "test code",
		},
		{
			name: "empty Description",
			input: ErrorTypeDetail{
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
