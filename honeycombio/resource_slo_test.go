package honeycombio

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"

	"github.com/stretchr/testify/require"

	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAccHoneycombioSLO_basic(t *testing.T) {
	dataset, sliAlias := sloAccTestSetup(t)
	slo := &honeycombio.SLO{}

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
		},
	})
}

// Checks to ensure that if an SLO was removed from Honeycomb outside of Terraform (UI or API)
// that it is detected and planned for recreation.
func TestAccHoneycombioSLO_RecreateOnNotFound(t *testing.T) {
	dataset, sliAlias := sloAccTestSetup(t)
	slo := &honeycombio.SLO{}

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
	slo := &honeycombio.SLO{}

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

func TestHoneycombSLO_MD(t *testing.T) {
	client := testAccClient(t)
	if client.IsClassic(context.Background()) {
		t.Skip("MD SLOs are not supported in classic")
	}
	dataset1, dataset2, mdSLI := mdSLOAccTestSetup(t)

	mdSLO := &honeycombio.SLO{}

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigSLO_md(dataset1.Slug, dataset2.Slug, mdSLI.Alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSLOExists(t, "honeycombio_slo.md_test", mdSLO),
					resource.TestCheckResourceAttr("honeycombio_slo.md_test", "name", "TestAcc MD SLO"),
					resource.TestCheckNoResourceAttr("honeycombio_slo.md_test", "dataset"),
					resource.TestCheckResourceAttr("honeycombio_slo.md_test", "datasets.#", "2"),
					resource.TestCheckTypeSetElemAttr("honeycombio_slo.md_test", "datasets.*", dataset1.Slug),
					resource.TestCheckTypeSetElemAttr("honeycombio_slo.md_test", "datasets.*", dataset2.Slug),
					resource.TestCheckResourceAttr("honeycombio_slo.md_test", "description", "integration test MD SLO"),
					resource.TestCheckResourceAttr("honeycombio_slo.md_test", "sli", mdSLI.Alias),
					resource.TestCheckResourceAttr("honeycombio_slo.md_test", "target_percentage", "99.95"),
					resource.TestCheckResourceAttr("honeycombio_slo.md_test", "time_period", "30"),
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
	}
	`, dataset, sliAlias)
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
	}
	`, dataset, sliAlias)
}

func testAccConfigSLO_md(dataset1Slug, dataset2Slug, sliAlias string) string {
	return fmt.Sprintf(`
	resource "honeycombio_slo" "md_test" {
		name              = "TestAcc MD SLO"
		description       = "integration test MD SLO"
		sli               = "%s"
		target_percentage = 99.95
		time_period       = 30
		datasets     	  = ["%s", "%s"]
	}
	`, sliAlias, dataset1Slug, dataset2Slug)
}

func testAccCheckSLOExists(t *testing.T, name string, slo *honeycombio.SLO) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("\"%s\" not found in state", name)
		}

		client := testAccClient(t)
		rslo, err := client.SLOs.Get(context.Background(), honeycombio.EnvironmentWideSlug, resourceState.Primary.ID)
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

	sli, err := c.DerivedColumns.Create(ctx, dataset, &honeycombio.DerivedColumn{
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

func mdSLOAccTestSetup(t *testing.T) (honeycombio.Dataset, honeycombio.Dataset, honeycombio.DerivedColumn) {
	t.Helper()

	ctx := context.Background()
	c := testAccClient(t)

	dataset1, err := c.Datasets.Create(ctx, &honeycombio.Dataset{
		Name:        "test." + acctest.RandString(8),
		Description: "test dataset 1",
	})
	require.NoError(t, err)

	dataset2, err := c.Datasets.Create(ctx, &honeycombio.Dataset{
		Name:        "test." + acctest.RandString(8),
		Description: "test dataset 2",
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		c.Datasets.Update(ctx, &honeycombio.Dataset{
			Slug: dataset1.Slug,
			Settings: honeycombio.DatasetSettings{
				DeleteProtected: helper.ToPtr(false),
			},
		})
		err = c.Datasets.Delete(ctx, dataset1.Slug)
		require.NoError(t, err)

		c.Datasets.Update(ctx, &honeycombio.Dataset{
			Slug: dataset2.Slug,
			Settings: honeycombio.DatasetSettings{
				DeleteProtected: helper.ToPtr(false),
			},
		})
		err = c.Datasets.Delete(ctx, dataset2.Slug)
		require.NoError(t, err)
	})

	sli, err := c.DerivedColumns.Create(ctx, honeycombio.EnvironmentWideSlug, &honeycombio.DerivedColumn{
		Alias:       test.RandomStringWithPrefix("test.", 10),
		Description: "test SLI",
		Expression:  "BOOL(1)",
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		c.DerivedColumns.Delete(ctx, honeycombio.EnvironmentWideSlug, sli.ID)
	})

	return *dataset1, *dataset2, *sli
}
