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

		assert.False(t, features.Client.ProactiveThrottling)
		assert.False(t, features.Column.ImportOnConflict)
		assert.False(t, features.Dataset.ImportOnConflict)
		assert.False(t, features.Intelligence.Enabled)
	})

	t.Run("parses client features", func(t *testing.T) {
		testCases := map[string]struct {
			model  []Model
			expect bool
		}{
			"parses ProactiveThrottling as false": {
				model: []Model{
					{
						Client: []FeaturesClientModel{
							{
								ProactiveThrottling: types.BoolValue(false),
							},
						},
					},
				},
				expect: false,
			},
			"parses ProactiveThrottling as true": {
				model: []Model{
					{
						Client: []FeaturesClientModel{
							{
								ProactiveThrottling: types.BoolValue(true),
							},
						},
					},
				},
				expect: true,
			},
			"handles Null": {
				model: []Model{
					{
						Client: []FeaturesClientModel{
							{
								ProactiveThrottling: types.BoolNull(),
							},
						},
					},
				},
				expect: false,
			},
			"handles Unknown": {
				model: []Model{
					{
						Client: []FeaturesClientModel{
							{
								ProactiveThrottling: types.BoolUnknown(),
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
				assert.Equal(t, tc.expect, features.Client.ProactiveThrottling)
			})
		}
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

	t.Run("parses dataset features", func(t *testing.T) {
		testCases := map[string]struct {
			model  []Model
			expect bool
		}{
			"parses ImportOnConflict as false": {
				model: []Model{
					{
						Dataset: []FeaturesDatasetModel{
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
						Dataset: []FeaturesDatasetModel{
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
						Dataset: []FeaturesDatasetModel{
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
						Dataset: []FeaturesDatasetModel{
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
				assert.Equal(t, tc.expect, features.Dataset.ImportOnConflict)
			})
		}
	})

	t.Run("parses client features from PluginSDK config", func(t *testing.T) {
		testCases := map[string]struct {
			raw    []any
			expect bool
		}{
			"handles empty features list": {
				raw:    []any{},
				expect: false,
			},
			"handles empty features block": {
				raw:    []any{nil},
				expect: false,
			},
			"handles missing client block": {
				raw:    []any{map[string]any{}},
				expect: false,
			},
			"handles empty client block": {
				raw: []any{map[string]any{
					"client": []any{},
				}},
				expect: false,
			},
			"parses ProactiveThrottling as false": {
				raw: []any{map[string]any{
					"client": []any{map[string]any{
						"proactive_throttling": false,
					}},
				}},
				expect: false,
			},
			"parses ProactiveThrottling as true": {
				raw: []any{map[string]any{
					"client": []any{map[string]any{
						"proactive_throttling": true,
					}},
				}},
				expect: true,
			},
		}

		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				features := ParsePluginSDK(tc.raw)
				assert.Equal(t, tc.expect, features.Client.ProactiveThrottling)
			})
		}
	})

	t.Run("parses intelligence features", func(t *testing.T) {
		testCases := map[string]struct {
			model  []Model
			expect bool
		}{
			"parses Enabled as false": {
				model: []Model{
					{
						Intelligence: []FeaturesIntelligenceModel{
							{
								Enabled: types.BoolValue(false),
							},
						},
					},
				},
				expect: false,
			},
			"parses Enabled as true": {
				model: []Model{
					{
						Intelligence: []FeaturesIntelligenceModel{
							{
								Enabled: types.BoolValue(true),
							},
						},
					},
				},
				expect: true,
			},
			"handles Null": {
				model: []Model{
					{
						Intelligence: []FeaturesIntelligenceModel{
							{
								Enabled: types.BoolNull(),
							},
						},
					},
				},
				expect: false,
			},
			"handles Unknown": {
				model: []Model{
					{
						Intelligence: []FeaturesIntelligenceModel{
							{
								Enabled: types.BoolUnknown(),
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
				assert.Equal(t, tc.expect, features.Intelligence.Enabled)
			})
		}
	})
}
