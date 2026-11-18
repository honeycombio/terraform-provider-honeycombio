package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	v2client "github.com/honeycombio/terraform-provider-honeycombio/client/v2"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAcc_GroupMembersResource(t *testing.T) {
	// the users are looked up before resource.Test, which would skip
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf("set %s to run this acceptance test", resource.EnvTfAcc)
	}
	users := testAccTeamUserIDs(t, 3)
	name := test.RandomStringWithPrefix("test.", 20)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
		CheckDestroy:             testAccEnsureGroupDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigGroupMembers(name, users[0], users[1]),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("honeycombio_group_members.test", "id", "honeycombio_group.test", "id"),
					resource.TestCheckResourceAttrPair("honeycombio_group_members.test", "group_id", "honeycombio_group.test", "id"),
					resource.TestCheckResourceAttr("honeycombio_group_members.test", "user_ids.#", "2"),
					testAccEnsureGroupMembersAre(t, users[0], users[1]),
				),
			},
			{ // changing the users is an update, not a replacement
				Config: testAccConfigGroupMembers(name, users[1], users[2]),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("honeycombio_group_members.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: testAccEnsureGroupMembersAre(t, users[1], users[2]),
			},
			{
				ResourceName:      "honeycombio_group_members.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // a user added outside Terraform is removed again
				PreConfig: func() {
					c := testAccV2Client(t)
					groupID := testAccGroupIDByName(t, name)
					if err := c.Groups.ReplaceMembers(context.Background(), groupID, users); err != nil {
						t.Fatal(err)
					}
				},
				Config: testAccConfigGroupMembers(name, users[1], users[2]),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("honeycombio_group_members.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: testAccEnsureGroupMembersAre(t, users[1], users[2]),
			},
			{ // a Group deleted outside Terraform is created again, with its members
				PreConfig: func() {
					groupID := testAccGroupIDByName(t, name)
					if err := testAccV2Client(t).Groups.Delete(context.Background(), groupID); err != nil {
						t.Fatal(err)
					}
				},
				Config: testAccConfigGroupMembers(name, users[1], users[2]),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("honeycombio_group.test", plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction("honeycombio_group_members.test", plancheck.ResourceActionCreate),
					},
				},
				Check: testAccEnsureGroupMembersAre(t, users[1], users[2]),
			},
			{ // an empty set removes every user
				Config: testAccConfigGroupMembers(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_group_members.test", "user_ids.#", "0"),
					testAccEnsureGroupMembersAre(t),
				),
			},
			{ // each user who is not on the Team gets its own error
				Config: testAccConfigGroupMembers(name, users[0], "hcxus_notonthisteam0000000", "hcxus_notonthisteam0000001"),
				ExpectError: regexp.MustCompile(
					`(?s)(hcxus_notonthisteam0000000: no user.*hcxus_notonthisteam0000001: no user|` +
						`hcxus_notonthisteam0000001: no user.*hcxus_notonthisteam0000000: no user)`,
				),
			},
			{ // add a user back, so the delete below has a user to remove
				Config: testAccConfigGroupMembers(name, users[0]),
				Check:  testAccEnsureGroupMembersAre(t, users[0]),
			},
			{ // deleting the resource removes every user and keeps the Group
				Config: testAccConfigGroup(name, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccEnsureGroupExists(t, "honeycombio_group.test"),
					testAccEnsureGroupMembersAre(t),
				),
			},
		},
	})
}

func testAccConfigGroupMembers(name string, userIDs ...string) string {
	quoted := make([]string, len(userIDs))
	for i, id := range userIDs {
		quoted[i] = fmt.Sprintf("%q", id)
	}
	return testAccConfigGroup(name, "") + fmt.Sprintf(`

resource "honeycombio_group_members" "test" {
  group_id = honeycombio_group.test.id
  user_ids = [%s]
}`, strings.Join(quoted, ", "))
}

// testAccTeamUserIDs returns the IDs of n users on the Team. The test fails
// when the Team has fewer, so CI can't pass without running it.
func testAccTeamUserIDs(t *testing.T, n int) []string {
	t.Helper()

	ctx := context.Background()
	pager, err := testAccV2Client(t).Users.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for pager.HasNext() && len(ids) < n {
		users, err := pager.Next(ctx)
		if err != nil {
			t.Fatal(err)
		}
		for _, u := range users {
			ids = append(ids, u.ID)
		}
	}
	if len(ids) < n {
		t.Fatalf("the Team needs at least %d users, and has %d", n, len(ids))
	}
	return ids[:n]
}

func testAccGroupIDByName(t *testing.T, name string) string {
	t.Helper()

	ctx := context.Background()
	pager, err := testAccV2Client(t).Groups.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for pager.HasNext() {
		groups, err := pager.Next(ctx)
		if err != nil {
			t.Fatal(err)
		}
		for _, g := range groups {
			if g.Name == name {
				return g.ID
			}
		}
	}
	t.Fatalf("no Group named %q", name)
	return ""
}

// testAccEnsureGroupMembersAre checks that the API lists exactly userIDs
// as the members of honeycombio_group.test.
func testAccEnsureGroupMembersAre(t *testing.T, userIDs ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["honeycombio_group.test"]
		if !ok {
			return fmt.Errorf("resource not found: honeycombio_group.test")
		}

		got, err := testAccGroupUserIDs(rs.Primary.ID)
		if err != nil {
			return err
		}
		want := slices.Clone(userIDs)
		slices.Sort(got)
		slices.Sort(want)
		if !slices.Equal(got, want) {
			return fmt.Errorf("Group members are %v in the API, want %v", got, want)
		}
		return nil
	}
}

// testAccGroupUserIDs returns the IDs of the users the API lists in the Group.
func testAccGroupUserIDs(groupID string) ([]string, error) {
	c, err := v2client.NewClient()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	pager, err := c.Groups.ListMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}
	userIDs := []string{}
	for pager.HasNext() {
		members, err := pager.Next(ctx)
		if err != nil {
			return nil, err
		}
		for _, m := range members {
			userIDs = append(userIDs, m.ID)
		}
	}
	return userIDs, nil
}
