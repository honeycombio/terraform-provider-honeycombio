package test

import (
	"bytes"
	"encoding/json"
	"math/rand"
)

const alphaNumericChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// RandomString generates a random string of the given length.
func RandomString(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = alphaNumericChars[rand.Intn(len(alphaNumericChars))]
	}
	return string(result)
}

// RandomStringWithPrefix generates a random string of characters of the given length
// with the provided prefix.
// This is useful for generating unique names for resources.
//
// Example:
//
//	RandomStringWithPrefix("test-", 10) => "test-abcde12345"
func RandomStringWithPrefix(prefix string, length int) string {
	return prefix + RandomString(length)
}

// MinifyJSON minifies a JSON string removing all whitespace and newlines
func MinifyJSON(s string) (string, error) {
	var buffer bytes.Buffer
	err := json.Compact(&buffer, []byte(s))
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}
