package client

import "reflect"

// Returns a pointer to the given value
func ToPtr[T any](v T) *T {
	return &v
}

// Determines if two slices of the same type are equivalent
// as opposed to equal
//
// For example: []string{"bob", "alice"} is equivalent but not equal to []string{"alice", "bob"}
func Equivalent[T any](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	for _, ours := range a {
		found := false
		for _, theirs := range b {
			if reflect.DeepEqual(ours, theirs) {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}
	// do it the other way around to make sure we're not missing a match
	for _, theirs := range b {
		found := false
		for _, ours := range a {
			if reflect.DeepEqual(theirs, ours) {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func IsZero[T comparable](v T) bool {
	return v == *new(T)
}

func PtrValueOrDefault[T any](v *T, d T) T {
	if v != nil {
		return *v
	}
	return d
}

func ValueOrDefault[T comparable](v, d T) T {
	if !IsZero(v) {
		return v
	}
	return d
}
