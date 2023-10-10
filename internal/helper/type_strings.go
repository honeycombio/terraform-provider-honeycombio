package helper

// AsStringSlice converts a slice of string-like things to a slice of strings.
func AsStringSlice[S ~string](in []S) []string {
	r := make([]string, len(in))
	for i := range in {
		r[i] = string(in[i])
	}

	return r
}
