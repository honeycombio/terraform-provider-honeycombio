package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEquivalent(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	tests := []struct {
		name string
		a, b []any
		want bool
	}{
		{"True - Identical", []any{"bob", "alice", "carol"}, []any{"bob", "alice", "carol"}, true},
		{"True - Out of order", []any{1, 2, 3}, []any{2, 3, 1}, true},
		{"True - Equiv Structs out of order", []any{Person{"Bob", 42}, Person{"Carol", 32}, Person{"Mallory", 25}}, []any{Person{"Mallory", 25}, Person{"Carol", 32}, Person{"Bob", 42}}, true},
		{"False - Not the same length", []any{1, 2, 3, 4}, []any{2, 3, 1}, false},
		{"False - Different values", []any{Person{"Bob", 42}, Person{"Carol", 32}, Person{"Mallory", 25}}, []any{Person{"Judy", 47}, Person{"Olivia", 27}, Person{"Frank", 21}}, false},
		{"False - Different cases", []any{"Bob", "Alice", "Carol"}, []any{"bob", "alice", "carol"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Equivalent(tt.a, tt.b))
		})
	}
}
