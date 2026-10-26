package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSummarize(t *testing.T) {
	t.Parallel()

	t.Run("lists only the innermost failing tests", func(t *testing.T) {
		t.Parallel()

		got := summarize(map[string]string{
			"TestParent":           "",
			"TestParent/child_bad": "    a_test.go:5: boom: x",
			"TestPanicky":          "",
			"TestOther/a":          "setup noise\nError: 429 Too Many Requests\nmore",
			"TestOther":            "",
		})
		assert.Equal(t, "• `TestOther/a`: Error: 429 Too Many Requests\n"+
			"• `TestPanicky`\n"+
			"• `TestParent/child_bad`: a_test.go:5: boom: x", got)
	})

	t.Run("explains a failure without failing tests", func(t *testing.T) {
		t.Parallel()

		assert.Contains(t, summarize(nil), "No failing tests were recorded")
	})
}
