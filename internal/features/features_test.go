package features

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestModelParse(t *testing.T) {
	t.Parallel()

	t.Run("handles empty model with defaults", func(t *testing.T) {
		model := []Model{}
		features := Parse(model)

		assert.False(t, features.Column.ImportOnConflict)
	})

	t.Run("parses column features", func(t *testing.T) {
		testCases := map[string]struct {
			model  []Model
			expect bool
		}{
			"parses ImportOnConflict as false": {
				model: []Model{
					{
						Column: []FeaturesColumnModel{
							{
								ImportOnConflict: types.BoolValue(false),
							},
						},
					},
				},
				expect: false,
			},
			"parses ImportOnConflict as true": {
				model: []Model{
					{
						Column: []FeaturesColumnModel{
							{
								ImportOnConflict: types.BoolValue(true),
							},
						},
					},
				},
				expect: true,
			},
			"handles Null": {
				model: []Model{
					{
						Column: []FeaturesColumnModel{
							{
								ImportOnConflict: types.BoolNull(),
							},
						},
					},
				},
				expect: false,
			},
			"handles Unknown": {
				model: []Model{
					{
						Column: []FeaturesColumnModel{
							{
								ImportOnConflict: types.BoolUnknown(),
							},
						},
					},
				},
				expect: false,
			},
		}

		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				features := Parse(tc.model)
				assert.Equal(t, tc.expect, features.Column.ImportOnConflict)
			})
		}
	})
}
