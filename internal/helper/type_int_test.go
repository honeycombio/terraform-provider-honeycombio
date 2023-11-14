package helper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypeInt_PPMToFloat(t *testing.T) {
	testCases := []struct {
		input    int
		expected float64
	}{
		{input: 1000000, expected: 100},
		{input: 990000, expected: 99},
		{input: 99999, expected: 9.9999},
		{input: 90000, expected: 9},
		{input: 1000, expected: 0.1},
		{input: 100, expected: 0.01},
		{input: 10, expected: 0.001},
		{input: 1, expected: 0.0001},
	}
	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("returns correct value for input %d", testCase.input), func(t *testing.T) {
			assert.Equal(t, testCase.expected, PPMToFloat(testCase.input))
		})
	}
}
