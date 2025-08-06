package features

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestModelParse(t *testing.T) {
	t.Run("parses column features correctly", func(t *testing.T) {
		model := Model{
			Column: []FeaturesColumnModel{
				{
					ImportOnConflict: types.BoolValue(true),
				},
			},
		}

		features := Parse(model)
		assert.True(t, features.Column.ImportOnConflict)
	})

	t.Run("handles empty model", func(t *testing.T) {
		model := Model{}
		features := Parse(model)

		assert.False(t, features.Column.ImportOnConflict)
	})

	t.Run("handles null and unknown values", func(t *testing.T) {
		model := Model{
			Column: []FeaturesColumnModel{
				{
					ImportOnConflict: types.BoolNull(),
				},
			},
		}

		features := Parse(model)
		assert.False(t, features.Column.ImportOnConflict)
	})

	t.Run("handles unknown values", func(t *testing.T) {
		model := Model{
			Column: []FeaturesColumnModel{
				{
					ImportOnConflict: types.BoolUnknown(),
				},
			},
		}

		features := Parse(model)

		assert.False(t, features.Column.ImportOnConflict)
	})
}
