package helper

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

// GetDatasetOrAll returns the dataset.
// If the dataset is empty, it returns the 'magic' EnvironmentWideSlug `__all__`.
func GetDatasetOrAll(dataset types.String) types.String {
	if dataset == types.StringNull() || dataset.ValueString() == "" {
		return types.StringValue(client.EnvironmentWideSlug)
	}
	return dataset
}
