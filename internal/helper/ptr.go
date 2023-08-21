package helper

// Returns a pointer to the given value
func ToPtr[T any](v T) *T {
	return &v
}
