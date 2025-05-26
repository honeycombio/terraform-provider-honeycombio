package helper

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
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

// DatasetSlugsToSet converts a slice of dataset slugs to a types.Set
// It returns a populated set if datasets exist, or a null set if no datasets are present
func DatasetSlugsToSet(ctx context.Context, datasetSlugs []string) (types.Set, diag.Diagnostics) {
	if len(datasetSlugs) > 0 {
		// Ensure we're not working with nil
		ds := datasetSlugs
		if ds == nil {
			ds = []string{}
		}
		return types.SetValueFrom(ctx, types.StringType, ds)
	}
	return types.SetNull(types.StringType), nil
}
