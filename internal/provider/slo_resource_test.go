package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAccHoneycombioSLO_basic(t *testing.T) {
	dataset, sliAlias := sloAccTestSetup(t)
	slo := &client.SLO{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "honeycombio_slo" "test" {
  name              = "TestAcc SLO"
  description       = "integration test SLO"
  dataset           = "%s"
  sli               = "%s"
  target_percentage = 99.95
  time_period       = 30

  tags = {
    env  = "test"
    team = "blue"
  }
}`, dataset, sliAlias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSLOExists(t, "honeycombio_slo.test", slo),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "dataset", dataset),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "name", "TestAcc SLO"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "description", "integration test SLO"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "sli", sliAlias),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "target_percentage", "99.95"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "time_period", "30"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "tags.%", "2"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "tags.env", "test"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "tags.team", "blue"),
				),
			},
			{ // update the config to remove the tags and change the description and percentage
				Config: fmt.Sprintf(`
resource "honeycombio_slo" "test" {
  name              = "TestAcc SLO"
  description       = "updated integration test SLO"
  dataset           = "%s"
  sli               = "%s"
  target_percentage = 99.99
  time_period       = 30
}`, dataset, sliAlias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSLOExists(t, "honeycombio_slo.test", slo),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "dataset", dataset),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "name", "TestAcc SLO"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "description", "updated integration test SLO"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "sli", sliAlias),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "target_percentage", "99.99"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "time_period", "30"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "tags.%", "0"),
				),
			},
			{ // test tags set to empty map
				Config: fmt.Sprintf(`
resource "honeycombio_slo" "test" {
  name              = "TestAcc SLO"
  description       = "updated integration test SLO"
  dataset           = "%s"
  sli               = "%s"
  target_percentage = 99.99
  time_period       = 30

  tags = {}
}`, dataset, sliAlias),
				Check: resource.TestCheckResourceAttr("honeycombio_slo.test", "tags.%", "0"),
			},
			{ // test tags set to null
				Config: fmt.Sprintf(`
resource "honeycombio_slo" "test" {
  name              = "TestAcc SLO"
  description       = "updated integration test SLO"
  dataset           = "%s"
  sli               = "%s"
  target_percentage = 99.99
  time_period       = 30

  tags = null
}`, dataset, sliAlias),
				Check: resource.TestCheckResourceAttr("honeycombio_slo.test", "tags.%", "0"),
			},
			{
				ResourceName:        "honeycombio_slo.test",
				ImportStateIdPrefix: fmt.Sprintf("%s/", dataset),
				ImportState:         true,
				ImportStateVerify:   true,
			},
		},
	})
}

// Checks to ensure that if an SLO was removed from Honeycomb outside of Terraform (UI or API)
// that it is detected and planned for recreation.
func TestAccHoneycombioSLO_RecreateOnNotFound(t *testing.T) {
	dataset, sliAlias := sloAccTestSetup(t)
	slo := &client.SLO{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigSLO_basic(dataset, sliAlias),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSLOExists(t, "honeycombio_slo.test", slo),
					func(_ *terraform.State) error {
						// the final 'check' deletes the SLO directly via the API leaving it behind in the state
						err := testAccClient(t).SLOs.Delete(context.Background(), dataset, slo.ID)
						if err != nil {
							return fmt.Errorf("failed to delete SLO: %w", err)
						}
						return nil
					},
				),
				// ensure that the plan is non-empty after the deletion
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccHoneycombioSLO_dataset_deprecation(t *testing.T) {
	dataset, sliAlias := sloAccTestSetup(t)
	slo := &client.SLO{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigSLO_basic(dataset, sliAlias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSLOExists(t, "honeycombio_slo.test", slo),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "dataset", dataset),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "name", "TestAcc SLO"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "description", "integration test SLO"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "sli", sliAlias),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "target_percentage", "99.95"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "time_period", "30"),
				),
			},
			// update the config to swap out dataset for datasets and ensure nothing changes
			{
				Config: testAccConfigSLO_dataset_deprecation(dataset, sliAlias),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_slo.test", "datasets.#", "1"),
					resource.TestCheckTypeSetElemAttr("honeycombio_slo.md_test", "datasets.*", dataset),
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccHoneycombSLO_MD(t *testing.T) {
	c := testAccClient(t)
	if c.IsClassic(context.Background()) {
		t.Skip("MD SLOs are not supported in classic")
	}
	dataset1, dataset2, mdSLI := mdSLOAccTestSetup(t)

	mdSLO := &client.SLO{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigSLO_md(dataset1.Slug, dataset2.Slug, mdSLI.Alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSLOExists(t, "honeycombio_slo.test", mdSLO),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "name", "TestAcc MD SLO"),
					resource.TestCheckNoResourceAttr("honeycombio_slo.test", "dataset"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "datasets.#", "2"),
					resource.TestCheckTypeSetElemAttr("honeycombio_slo.test", "datasets.*", dataset1.Slug),
					resource.TestCheckTypeSetElemAttr("honeycombio_slo.test", "datasets.*", dataset2.Slug),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "description", "integration test MD SLO"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "sli", mdSLI.Alias),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "target_percentage", "99.95"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "time_period", "30"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "tags.%", "2"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "tags.env", "test"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "tags.team", "red"),
				),
			},
			// tests imports
			{
				ResourceName:      "honeycombio_slo.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     mdSLO.ID,
			},
		},
	})
}

func TestAccHoneycombioSLO_UpgradeFromSDK(t *testing.T) {
	dataset, sliAlias := sloAccTestSetup(t)
	slo := &client.SLO{}

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"honeycombio": {
						VersionConstraint: "0.35.0",
						Source:            "honeycombio/honeycombio",
					},
				},
				Config: fmt.Sprintf(`
					resource "honeycombio_slo" "test" {
						name              = "TestAcc SLO Upgrade"
						description       = "integration test SLO for SDK upgrade"
						dataset           = "%s"
						sli               = "%s" 
						target_percentage = 99.95
						time_period       = 30

						tags = {
							env  = "test"
							team = "upgrade"
						}
					}`,
					dataset, sliAlias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSLOExists(t, "honeycombio_slo.test", slo),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "dataset", dataset),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "name", "TestAcc SLO Upgrade"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "target_percentage", "99.95"),
				),
			},
			{
				ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
				Config: fmt.Sprintf(`
					resource "honeycombio_slo" "test" {
						name              = "TestAcc SLO Upgrade" 
						description       = "integration test SLO for SDK upgrade"
						dataset           = "%s"
						sli               = "%s"
						target_percentage = 99.95
						time_period       = 30

						tags = {
							env  = "test"
							team = "upgrade" 
						}
					}`,
					dataset, sliAlias),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccHoneycombioSLO_DatasetConstraint(t *testing.T) {
	dataset, sliAlias := sloAccTestSetup(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "honeycombio_slo" "test" {
						name              = "TestAcc SLO Constraint"
						description       = "integration test SLO constraint"
						dataset           = "%s"
						datasets          = ["%s"]
						sli               = "%s"
						target_percentage = 99.95
						time_period       = 30
					}`,
					dataset, dataset, sliAlias),
				ExpectError: regexp.MustCompile(`"datasets" cannot be specified when "dataset" is specified`),
			},
			{
				Config: fmt.Sprintf(`
				resource "honeycombio_slo" "test" {
					name              = "TestAcc SLO Constraint"
					description       = "integration test SLO constraint"
					sli               = "%s" 
					target_percentage = 99.95
					time_period       = 30
					}`, sliAlias),
				ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
			},
		},
	})
}

func TestAccHoneycombioSLO_Update(t *testing.T) {
	c := testAccClient(t)
	if c.IsClassic(context.Background()) {
		t.Skip("Multi-dataset SLOs are not supported in classic")
	}

	dataset1, dataset2, sliAlias := mdSLOAccTestSetup(t)

	slo := &honeycombio.SLO{}
	var originalID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "honeycombio_slo" "test" {
  name              = "TestAcc SLO"
  description       = "test SLO"
  datasets           = ["%s"]
  sli               = "%s"
  target_percentage = 99.95
  time_period       = 30

  tags = {
    env  = "test"
    team = "blue"
  }
}`, dataset1.Slug, sliAlias.Alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSLOExists(t, "honeycombio_slo.test", slo),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["honeycombio_slo.test"]
						originalID = rs.Primary.ID
						return nil
					},
					resource.TestCheckResourceAttr("honeycombio_slo.test", "name", "TestAcc SLO"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "description", "test SLO"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "target_percentage", "99.95"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "time_period", "30"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "datasets.#", "1"),
					resource.TestCheckTypeSetElemAttr("honeycombio_slo.test", "datasets.*", dataset1.Slug),
				),
			},
			{ // Update description - should not change ID
				Config: fmt.Sprintf(`
resource "honeycombio_slo" "test" {
  name              = "TestAcc SLO"
  description       = "UPDATED integration test SLO"
  datasets           = ["%s"]
  sli               = "%s"
  target_percentage = 99.95
  time_period       = 30

  tags = {
    env  = "test"
    team = "blue"
  }
}`, dataset1.Slug, sliAlias.Alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSLOExists(t, "honeycombio_slo.test", slo),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["honeycombio_slo.test"]
						if rs.Primary.ID != originalID {
							return fmt.Errorf("expected ID to remain %s after description update, but got %s", originalID, rs.Primary.ID)
						}
						return nil
					},
					resource.TestCheckResourceAttr("honeycombio_slo.test", "description", "UPDATED integration test SLO"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "name", "TestAcc SLO"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "target_percentage", "99.95"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "time_period", "30"),
					resource.TestCheckResourceAttr("honeycombio_slo.test", "datasets.#", "1"),
					resource.TestCheckTypeSetElemAttr("honeycombio_slo.test", "datasets.*", dataset1.Slug),
				),
			},
			{ // Update datasets - should CHANGE ID (force replacement)
				Config: fmt.Sprintf(`
resource "honeycombio_slo" "test" {
  name              = "TestAcc SLO Update"
  description       = "UPDATED test SLO for updates"
  datasets          = ["%s"]
  sli               = "%s"
  target_percentage = 99.99
  time_period       = 30

  tags = {
    env  = "production"
    team = "red"
    owner = "devops"
  }
}`, dataset2.Slug, sliAlias.Alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSLOExists(t, "honeycombio_slo.test", slo),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["honeycombio_slo.test"]
						if rs.Primary.ID == originalID {
							return fmt.Errorf("expected ID to change after datasets update, but it remained %s", originalID)
						}
						originalID = rs.Primary.ID
						return nil
					},
					resource.TestCheckResourceAttr("honeycombio_slo.test", "datasets.#", "1"),
					resource.TestCheckTypeSetElemAttr("honeycombio_slo.test", "datasets.*", dataset2.Slug),
				),
			},
		},
	})
}

func testAccConfigSLO_basic(dataset, sliAlias string) string {
	return fmt.Sprintf(`
resource "honeycombio_slo" "test" {
  name              = "TestAcc SLO"
  description       = "integration test SLO"
  dataset           = "%s"
  sli               = "%s"
  target_percentage = 99.95
  time_period       = 30

  tags = {
    env  = "test"
    team = "blue"
  }
}`, dataset, sliAlias)
}

func testAccConfigSLO_dataset_deprecation(dataset, sliAlias string) string {
	return fmt.Sprintf(`
resource "honeycombio_slo" "test" {
  name              = "TestAcc SLO"
  description       = "integration test SLO"
  datasets          = ["%s"]
  sli               = "%s"
  target_percentage = 99.95
  time_period       = 30

  tags = {
    env  = "test"
    team = "blue"
  }
}`, dataset, sliAlias)
}

func testAccConfigSLO_md(dataset1Slug, dataset2Slug, sliAlias string) string {
	return fmt.Sprintf(`
resource "honeycombio_slo" "test" {
  name              = "TestAcc MD SLO"
  description       = "integration test MD SLO"
  sli               = "%s"
  datasets     	    = ["%s", "%s"]
  target_percentage = 99.95
  time_period       = 30

  tags = {
    env  = "test"
    team = "red"
  }
}`, sliAlias, dataset1Slug, dataset2Slug)
}

func testAccCheckSLOExists(t *testing.T, name string, slo *client.SLO) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("\"%s\" not found in state", name)
		}

		c := testAccClient(t)
		rslo, err := c.SLOs.Get(context.Background(), client.EnvironmentWideSlug, resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch created SLO: %w", err)
		}

		*slo = *rslo

		return nil
	}
}

func sloAccTestSetup(t *testing.T) (string, string) {
	t.Helper()

	ctx := context.Background()
	c := testAccClient(t)
	dataset := testAccDataset()

	sli, err := c.DerivedColumns.Create(ctx, dataset, &client.DerivedColumn{
		Alias:      "sli." + acctest.RandString(8),
		Expression: "BOOL(1)",
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		// remove SLI DC at end of test run
		c.DerivedColumns.Delete(ctx, dataset, sli.ID)
	})

	return dataset, sli.Alias
}

func mdSLOAccTestSetup(t *testing.T) (client.Dataset, client.Dataset, client.DerivedColumn) {
	t.Helper()

	ctx := context.Background()
	c := testAccClient(t)

	dataset1, err := c.Datasets.Create(ctx, &client.Dataset{
		Name:        "test." + acctest.RandString(8),
		Description: "test dataset 1",
	})
	require.NoError(t, err)

	dataset2, err := c.Datasets.Create(ctx, &client.Dataset{
		Name:        "test." + acctest.RandString(8),
		Description: "test dataset 2",
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		c.Datasets.Update(ctx, &client.Dataset{
			Slug: dataset1.Slug,
			Settings: client.DatasetSettings{
				DeleteProtected: helper.ToPtr(false),
			},
		})
		err = c.Datasets.Delete(ctx, dataset1.Slug)
		require.NoError(t, err)

		c.Datasets.Update(ctx, &client.Dataset{
			Slug: dataset2.Slug,
			Settings: client.DatasetSettings{
				DeleteProtected: helper.ToPtr(false),
			},
		})
		err = c.Datasets.Delete(ctx, dataset2.Slug)
		require.NoError(t, err)
	})

	sli, err := c.DerivedColumns.Create(ctx, client.EnvironmentWideSlug, &client.DerivedColumn{
		Alias:       test.RandomStringWithPrefix("test.", 10),
		Description: "test SLI",
		Expression:  "BOOL(1)",
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		c.DerivedColumns.Delete(ctx, client.EnvironmentWideSlug, sli.ID)
	})

	return *dataset1, *dataset2, *sli
}
