package verify

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

func SupressEquivQuerySpecDiff(_, orig, new string, _ *schema.ResourceData) bool {
	var qs1, qs2 client.QuerySpec

	if err := json.Unmarshal([]byte(orig), &qs1); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &qs2); err != nil {
		return false
	}
	return qs1.EquivalentTo(qs2)
}
