package provider

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_UserDataSource(t *testing.T) {
	user := testAccTeamUsers(t, 1)[0]

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigUserDataSource(user.Email),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.honeycombio_user.test", "id", user.ID),
					resource.TestCheckResourceAttr("data.honeycombio_user.test", "email", user.Email),
					resource.TestCheckResourceAttr("data.honeycombio_user.test", "role", user.Role),
				),
			},
			{ // the email matches without regard to case
				Config: testAccConfigUserDataSource(strings.ToUpper(user.Email)),
				Check:  resource.TestCheckResourceAttr("data.honeycombio_user.test", "id", user.ID),
			},
			{
				Config:      testAccConfigUserDataSource("nobody.at.all@example.com"),
				ExpectError: regexp.MustCompile(`No user on the Team has the email "nobody.at.all@example.com"`),
			},
		},
	})
}

func TestUserDataSource_rejectsANamedEmail(t *testing.T) {
	// the validator fails the plan before the provider calls the API
	testAccFakeUsersAPI(t, testNotFound)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		Steps: []resource.TestStep{{
			Config:      testAccConfigUserDataSource("Alice <alice@example.com>"),
			ExpectError: regexp.MustCompile(`must be a bare email address`),
		}},
	})
}

func testAccConfigUserDataSource(email string) string {
	return fmt.Sprintf(`
data "honeycombio_user" "test" {
  email = "%s"
}`, email)
}
