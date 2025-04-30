package client

import "regexp"

// MaxTagsPerResource is the maximum number of tags that can be associated with a resource.
const MaxTagsPerResource = 10

var (
	// TagKeyValidationRegex is the regex used to validate tag keys.
	// It must be a string of lowercase letters, between 1 and 32 characters long.
	TagKeyValidationRegex = regexp.MustCompile(`^[a-z]{1,32}$`)

	// TagValueValidationRegex is the regex used to validate tag values.
	// It must begin with a lowercase letter, be between 1 and 32 characters long,
	// and only contain alphanumeric characters, -, or /.
	TagValueValidationRegex = regexp.MustCompile(`^[a-z][a-z0-9\/-]{1,32}$`)
)

// Tag represents a key-value pair used for tagging resources.
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
