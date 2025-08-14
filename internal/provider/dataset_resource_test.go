package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAcc_DatasetResource(t *testing.T) {

	t.Run("happy path", func(t *testing.T) {
		name := test.RandomStringWithPrefix("test.", 20)
		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheck(t),
			ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
			Steps: []resource.TestStep{
				{
					Config: testAccConfigBasicDatasetTest(name, "test description", 0, true),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureDatasetExists(t, "honeycombio_dataset.test"),
						resource.TestCheckResourceAttrSet("honeycombio_dataset.test", "id"),
						resource.TestCheckResourceAttrSet("honeycombio_dataset.test", "slug"),
						resource.TestCheckResourceAttr("honeycombio_dataset.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_dataset.test", "description", "test description"),
						resource.TestCheckResourceAttr("honeycombio_dataset.test", "expand_json_depth", "0"),
						resource.TestCheckResourceAttr("honeycombio_dataset.test", "delete_protected", "true"),
						resource.TestCheckResourceAttrSet("honeycombio_dataset.test", "created_at"),
						resource.TestCheckResourceAttrSet("honeycombio_dataset.test", "last_written_at"),
					),
				},
				{ // fail to delete protected dataset
					Config:      testAccConfigBasicDatasetTest(name, "test description", 0, true),
					Destroy:     true,
					ExpectError: regexp.MustCompile(`must disable delete protection`),
				},
				{ // now update the description and JSON expansion, and disable delete protection
					Config: testAccConfigBasicDatasetTest(name, "new description", 3, false),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureDatasetExists(t, "honeycombio_dataset.test"),
						resource.TestCheckResourceAttrSet("honeycombio_dataset.test", "id"),
						resource.TestCheckResourceAttrSet("honeycombio_dataset.test", "slug"),
						resource.TestCheckResourceAttr("honeycombio_dataset.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_dataset.test", "description", "new description"),
						resource.TestCheckResourceAttr("honeycombio_dataset.test", "expand_json_depth", "3"),
						resource.TestCheckResourceAttr("honeycombio_dataset.test", "delete_protected", "false"),
						resource.TestCheckResourceAttrSet("honeycombio_dataset.test", "created_at"),
						resource.TestCheckResourceAttrSet("honeycombio_dataset.test", "last_written_at"),
					),
				},
				{
					ResourceName: "honeycombio_dataset.test",
					ImportState:  true,
				},
			},
		})
	})

	t.Run("default values", func(t *testing.T) {
		name := test.RandomStringWithPrefix("test.", 20)
		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheck(t),
			ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
resource "honeycombio_dataset" "test" {
  name = "%s"
}`, name),
					Check: resource.ComposeAggregateTestCheckFunc(
						testAccEnsureDatasetExists(t, "honeycombio_dataset.test"),
						resource.TestCheckResourceAttrSet("honeycombio_dataset.test", "id"),
						resource.TestCheckResourceAttrSet("honeycombio_dataset.test", "slug"),
						resource.TestCheckResourceAttr("honeycombio_dataset.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_dataset.test", "description", ""),
						resource.TestCheckResourceAttr("honeycombio_dataset.test", "expand_json_depth", "0"),
						resource.TestCheckResourceAttr("honeycombio_dataset.test", "delete_protected", "true"),
						resource.TestCheckResourceAttrSet("honeycombio_dataset.test", "created_at"),
						resource.TestCheckResourceAttrSet("honeycombio_dataset.test", "last_written_at"),
					),
				},
				{ // disable delete protection to allow cleanup
					Config: fmt.Sprintf(`
resource "honeycombio_dataset" "test" {
  name             = "%s"
  delete_protected = false
}`, name),
				},
			},
		})
	})

	t.Run("fails to create dataset with delete protection disabled", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheck(t),
			ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
			Steps: []resource.TestStep{
				{
					Config:      testAccConfigBasicDatasetTest("nope", "", 0, false),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`protection cannot be disabled at creation`),
				},
			},
		})
	})

	t.Run("fails to create a dataset with duplicate name", func(t *testing.T) {
		ctx := context.Background()
		c := testAccClient(t)

		ds, err := c.Datasets.Create(ctx, &client.Dataset{
			Name: test.RandomStringWithPrefix("test.", 20),
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			// disable deletion protection and delete the Dataset
			c.Datasets.Update(ctx, &client.Dataset{
				Slug: ds.Slug,
				Settings: client.DatasetSettings{
					DeleteProtected: helper.ToPtr(false),
				},
			})
			c.Datasets.Delete(ctx, ds.Slug)
		})

		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheckV2API(t),
			ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
			Steps: []resource.TestStep{
				{
					Config:      testAccConfigBasicDatasetTest(ds.Name, "", 0, true),
					ExpectError: regexp.MustCompile(`Dataset "` + ds.Name + `" already exists`),
				},
			},
		})
	})

	t.Run("feature: import_on_conflict", func(t *testing.T) {
		ctx := context.Background()
		c := testAccClient(t)

		ds, err := c.Datasets.Create(ctx, &client.Dataset{
			Name:            test.RandomStringWithPrefix("test.", 20),
			Description:     "my dataset is nice",
			ExpandJSONDepth: 3,
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			// disable deletion protection and delete the Dataset
			c.Datasets.Update(ctx, &client.Dataset{
				Slug: ds.Slug,
				Settings: client.DatasetSettings{
					DeleteProtected: helper.ToPtr(false),
				},
			})
			c.Datasets.Delete(ctx, ds.Slug)
		})

		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheck(t),
			ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
			Steps: []resource.TestStep{
				{ // explicitly set import_on_conflict to false to ensure it fails
					Config: fmt.Sprintf(`
provider "honeycombio" {
  features {
    dataset {
      import_on_conflict = false
    }
  }
}

resource "honeycombio_dataset" "test" {
  name              = "%s"
  description       = "my dataset is nice"
  expand_json_depth = 3
}`, ds.Name),
					ExpectError: regexp.MustCompile(`Dataset "` + ds.Name + `" already exists`),
				},
				{ // set import_on_conflict to true to ensure it imports and updates the existing dataset
					Config: fmt.Sprintf(`
provider "honeycombio" {
  features {
    dataset {
      import_on_conflict = true
    }
  }
}

resource "honeycombio_dataset" "test" {
  name              = "%s"
  description       = "my dataset is imported"
  expand_json_depth = 5
}`, ds.Name),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("honeycombio_dataset.test", "id"),
						resource.TestCheckResourceAttrSet("honeycombio_dataset.test", "slug"),
						resource.TestCheckResourceAttr("honeycombio_dataset.test", "name", ds.Name),
						resource.TestCheckResourceAttr("honeycombio_dataset.test", "description", "my dataset is imported"), // updated description
						resource.TestCheckResourceAttr("honeycombio_dataset.test", "expand_json_depth", "5"),                // updated JSON depth
						resource.TestCheckResourceAttr("honeycombio_dataset.test", "delete_protected", "true"),
						resource.TestCheckResourceAttrSet("honeycombio_dataset.test", "created_at"),
						resource.TestCheckResourceAttrSet("honeycombio_dataset.test", "last_written_at"),
					),
				},
			},
		})
	})
}

// TestAcc_DatasetResource_UpgradeFromVersion026 tests the migration case from the
// last SDK-based version of the Dataset resource to the current Framework-based version.
//
// See: https://developer.hashicorp.com/terraform/plugin/framework/migrating/testing#testing-migration
func TestAcc_DatasetResource_UpgradeFromVersion026(t *testing.T) {
	name := test.RandomStringWithPrefix("test.", 20)

	resource.Test(t, resource.TestCase{
		PreCheck: testAccPreCheck(t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"honeycombio": {
						VersionConstraint: "0.26.0",
						Source:            "honeycombio/honeycombio",
					},
				},
				Config: fmt.Sprintf(`
resource "honeycombio_dataset" "test" {
  name = "%s"
}`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccEnsureDatasetExists(t, "honeycombio_dataset.test"),
				),
			},
			{
				ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
				Config: fmt.Sprintf(`
resource "honeycombio_dataset" "test" {
  name = "%s"
}`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("honeycombio_dataset.test", "id"),
					resource.TestCheckResourceAttrSet("honeycombio_dataset.test", "slug"),
					resource.TestCheckResourceAttr("honeycombio_dataset.test", "name", name),
					resource.TestCheckResourceAttr("honeycombio_dataset.test", "description", ""),
					resource.TestCheckResourceAttr("honeycombio_dataset.test", "expand_json_depth", "0"),
					resource.TestCheckResourceAttr("honeycombio_dataset.test", "delete_protected", "true"),
					resource.TestCheckResourceAttrSet("honeycombio_dataset.test", "created_at"),
					resource.TestCheckResourceAttrSet("honeycombio_dataset.test", "last_written_at"),
				),
			},
			{ // disable delete protection to allow cleanup
				ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
				Config: fmt.Sprintf(`
resource "honeycombio_dataset" "test" {
  name             = "%s"
  delete_protected = false
}`, name),
			},
		},
	})
}

func testAccConfigBasicDatasetTest(name, description string, jsonDepth int, protected bool) string {
	return fmt.Sprintf(`
resource "honeycombio_dataset" "test" {
  name        = "%s"
  description = "%s"

  expand_json_depth = %d
  delete_protected  = %t
}`, name, description, jsonDepth, protected)
}

func testAccEnsureDatasetExists(t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		client := testAccClient(t)
		_, err := client.Datasets.Get(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch created Dataset: %s", err)
		}

		return nil
	}
}
