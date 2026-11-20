package provider

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAcc_GroupDataSource(t *testing.T) {
	name := test.RandomStringWithPrefix("test.", 20)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigGroupDataSource(name, "honeycombio_group.test.name"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.honeycombio_group.test", "id", "honeycombio_group.test", "id"),
					resource.TestCheckResourceAttr("data.honeycombio_group.test", "name", name),
					resource.TestCheckResourceAttr("data.honeycombio_group.test", "description", "Reads payments"),
				),
			},
			{ // the same name in a different case does not match
				Config:      testAccConfigGroupDataSource(name, fmt.Sprintf("%q", strings.ToUpper(name))),
				ExpectError: regexp.MustCompile(`No Group on the Team has the name`),
			},
		},
	})
}

func testAccConfigGroupDataSource(groupName, lookupName string) string {
	return fmt.Sprintf(`
resource "honeycombio_group" "test" {
  name        = "%s"
  description = "Reads payments"
}

data "honeycombio_group" "test" {
  name = %s
}`, groupName, lookupName)
}
