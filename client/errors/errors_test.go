package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	t.Parallel()

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
