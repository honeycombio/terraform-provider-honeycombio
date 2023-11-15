package helper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypeFloat_FloatToPPM(t *testing.T) {
	testCases := []struct {
		input    float64
		expected int
	}{
		{input: 100, expected: 1000000},
		{input: 99.00, expected: 990000},
		{input: 9.9999, expected: 99999},
		{input: 9, expected: 90000},
		{input: 0.1, expected: 1000},
		{input: 0.010, expected: 100},
		{input: 0.001, expected: 10},
		{input: 0.0001, expected: 1},
	}
	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("returns correct value for input %g", testCase.input), func(t *testing.T) {
			assert.Equal(t, testCase.expected, FloatToPPM(testCase.input))
		})
	}
}

func TestTypeFloat_FloatToPercentString(t *testing.T) {
	testCases := []struct {
		input    float64
		expected string
	}{
		{input: 100, expected: "100"},
		{input: 99.00, expected: "99"},
		{input: 9.9999, expected: "9.9999"},
		{input: 9, expected: "9"},
		{input: 0.1, expected: "0.1"},
		{input: 0.010, expected: "0.01"},
		{input: 0.001, expected: "0.001"},
		{input: 0.0001, expected: "0.0001"},
	}
	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("returns correct value for input %g", testCase.input), func(t *testing.T) {
			assert.Equal(t, testCase.expected, FloatToPercentString(testCase.input))
		})
	}
}
