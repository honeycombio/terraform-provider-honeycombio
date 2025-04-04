package verify

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// SuppressEquivEnvWideDataset suppresses the dataset field if the old value is
// equivalent to the new value, or if the new value is empty.
//
// This is used to allow migrations from the magic '__all__' to the empty string
// to be handled gracefully, and not force re-creation of the resource.
func SuppressEquivEnvWideDataset(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	if oldValue == newValue {
		return true
	}
	// if the config moves away from deprecated dataset, nothing should change
	if newValue == "" && oldValue == "__all__" {
		return true
	}
	return false
}
