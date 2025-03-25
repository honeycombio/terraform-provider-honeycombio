package helper

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

func GetDatasetString(dataset types.String) types.String {
	if dataset == types.StringNull() {
		return types.StringValue(client.EnvironmentWideSlug)
	}
	return dataset
}
