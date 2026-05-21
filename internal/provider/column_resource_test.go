package provider

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"testing"
	"time"

	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/models"
)

func TestAcc_ColumnResource(t *testing.T) {
	ctx := context.Background()

	t.Run("happy path", func(t *testing.T) {
		dataset := testAccDataset()
		name := test.RandomStringWithPrefix("test.", 10)

		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheck(t),
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
resource "honeycombio_column" "test" {
  name        = "%s"
  type        = "float"
  hidden      = false
  description = "Duration of the trace"

  dataset = "%s"
}`, name, dataset),
					Check: resource.ComposeTestCheckFunc(
						testAccEnsureColumnExists(t, "honeycombio_column.test", name),
						resource.TestCheckResourceAttrSet("honeycombio_column.test", "id"),
						resource.TestCheckResourceAttr("honeycombio_column.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_column.test", "dataset", dataset),
						resource.TestCheckResourceAttr("honeycombio_column.test", "type", "float"),
						resource.TestCheckResourceAttr("honeycombio_column.test", "hidden", "false"),
						resource.TestCheckResourceAttr("honeycombio_column.test", "description", "Duration of the trace"),
						resource.TestCheckResourceAttrSet("honeycombio_column.test", "created_at"),
						resource.TestCheckResourceAttrSet("honeycombio_column.test", "updated_at"),
						resource.TestCheckResourceAttrSet("honeycombio_column.test", "last_written_at"),
					),
				},
				{
					SkipFunc: func() (bool, error) {
						t.Skip("column updates are racey and not reliably testable right now")
						return true, nil
					},
					Config: fmt.Sprintf(`
resource "honeycombio_column" "test" {
  name        = "%s"
  dataset     = "%s"
  type        = "float"
  hidden      = true
  description = "My nice column"
}`, name, dataset),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("honeycombio_column.test", "id"),
						resource.TestCheckResourceAttr("honeycombio_column.test", "name", name),
						resource.TestCheckResourceAttr("honeycombio_column.test", "dataset", dataset),
						resource.TestCheckResourceAttr("honeycombio_column.test", "type", "float"),
						resource.TestCheckResourceAttr("honeycombio_column.test", "hidden", "true"),
						resource.TestCheckResourceAttr("honeycombio_column.test", "description", "My nice column"),
						resource.TestCheckResourceAttrSet("honeycombio_column.test", "created_at"),
						resource.TestCheckResourceAttrSet("honeycombio_column.test", "updated_at"),
						resource.TestCheckResourceAttrSet("honeycombio_column.test", "last_written_at"),
					),
				},
				{
					ResourceName:      "honeycombio_column.test",
					ImportStateId:     fmt.Sprintf("%s/%s", dataset, name),
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})

	t.Run("feature: import_on_conflict", func(t *testing.T) {
		t.Parallel() // we don't want this test to block others as it sleeps for a while

		c := testAccClient(t)
		dataset := testAccDataset()

		// make a column so we can 'create' it and test the import_on_conflict behavior
		column, err := c.Columns.Create(ctx, dataset, &client.Column{
			KeyName: test.RandomStringWithPrefix("test.", 10),
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			c.Columns.Delete(ctx, dataset, column.KeyName)
		})

		// give the backend a chance to catch up
		time.Sleep(31 * time.Second)

		// column creation can be a bit racey, so we'll wait for it to be available
		assert.Eventually(t, func() bool {
			_, err := c.Columns.GetByKeyName(ctx, dataset, column.KeyName)
			return err == nil
		}, 5*time.Second, 200*time.Millisecond)

		resource.Test(t, resource.TestCase{
			PreCheck:                 testAccPreCheck(t),
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactory,
			Steps: []resource.TestStep{
				{ // explicitly set import_on_conflict to false to ensure it fails
					Config: fmt.Sprintf(`
provider "honeycombio" {
  features {
    column {
      import_on_conflict = false
    }
  }
}

resource "honeycombio_column" "test" {
  name        = "%s"
  dataset     = "%s"
}`, column.KeyName, dataset),
					ExpectError: regexp.MustCompile(`(?i)column already exists`),
				},
				{ // set import_on_conflict to true to ensure it imports the existing column
					Config: fmt.Sprintf(`
provider "honeycombio" {
  features {
    column {
      import_on_conflict = true
    }
  }
}

resource "honeycombio_column" "test" {
  name        = "%s"
  dataset     = "%s"
}`, column.KeyName, dataset),
					Check: resource.ComposeTestCheckFunc(
						testAccEnsureColumnExists(t, "honeycombio_column.test", column.KeyName),
						resource.TestCheckResourceAttrSet("honeycombio_column.test", "id"),
						resource.TestCheckResourceAttr("honeycombio_column.test", "name", column.KeyName),
						resource.TestCheckResourceAttr("honeycombio_column.test", "dataset", dataset),
						resource.TestCheckResourceAttr("honeycombio_column.test", "type", "string"), // default type is string
						resource.TestCheckResourceAttr("honeycombio_column.test", "hidden", "false"),
						resource.TestCheckResourceAttr("honeycombio_column.test", "description", ""),
						resource.TestCheckResourceAttrSet("honeycombio_column.test", "created_at"),
						resource.TestCheckResourceAttrSet("honeycombio_column.test", "updated_at"),
						resource.TestCheckResourceAttrSet("honeycombio_column.test", "last_written_at"),
					),
				},
			},
		})
	})
}

// TestAcc_ColumnResourceUpgradeFromVersion037 is intended to test the migration
// case from the last SDK-based version of the Column resource to the current Framework-based
// version.
//
// See: https://developer.hashicorp.com/terraform/plugin/framework/migrating/testing#testing-migration
func TestAcc_ColumnResourceUpgradeFromVersion037(t *testing.T) {
	dataset := testAccDataset()
	name := test.RandomStringWithPrefix("test.", 10)

	config := fmt.Sprintf(`
resource "honeycombio_column" "test" {
  name        = "%s"
  dataset     = "%s"
  description = "My nice column"
}`, name, dataset)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"honeycombio": {
						VersionConstraint: "0.37.1",
						Source:            "honeycombio/honeycombio",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccEnsureColumnExists(t, "honeycombio_column.test", name),
				),
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6MuxServerFactory,
				Config:                   config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// mockColumns is a minimal stub of client.Columns for unit tests.
type mockColumns struct {
	deleteErr error
}

func (m mockColumns) List(_ context.Context, _ string) ([]client.Column, error)  { return nil, nil }
func (m mockColumns) Get(_ context.Context, _, _ string) (*client.Column, error) { return nil, nil }
func (m mockColumns) GetByKeyName(_ context.Context, _, _ string) (*client.Column, error) {
	return nil, nil
}
func (m mockColumns) Create(_ context.Context, _ string, c *client.Column) (*client.Column, error) {
	return c, nil
}
func (m mockColumns) Update(_ context.Context, _ string, c *client.Column) (*client.Column, error) {
	return c, nil
}
func (m mockColumns) Delete(_ context.Context, _, _ string) error { return m.deleteErr }

func Test_columnResource_Delete(t *testing.T) {
	ctx := context.Background()

	cr := &columnResource{}
	var schemaResp tfresource.SchemaResponse
	cr.Schema(ctx, tfresource.SchemaRequest{}, &schemaResp)

	state := tfsdk.State{Schema: schemaResp.Schema}
	diags := state.Set(ctx, models.ColumnResourceModel{
		ID:            fwtypes.StringValue("col-123"),
		Dataset:       fwtypes.StringValue("my-dataset"),
		Name:          fwtypes.StringValue("duration_ms"),
		Hidden:        fwtypes.BoolValue(false),
		Description:   fwtypes.StringValue(""),
		Type:          fwtypes.StringValue("float"),
		CreatedAt:     fwtypes.StringValue(""),
		UpdatedAt:     fwtypes.StringValue(""),
		LastWrittenAt: fwtypes.StringValue(""),
	})
	require.False(t, diags.HasError(), "state setup failed: %v", diags)

	tests := []struct {
		name        string
		deleteErr   error
		wantErrDiag bool
	}{
		{
			name: "succeeds when column is in use by dataset definition",
			deleteErr: client.DetailedError{
				Status:  http.StatusConflict,
				Message: "Column is in use by dataset definition - duration_ms",
			},
			wantErrDiag: false,
		},
		{
			name:        "succeeds when delete succeeds",
			deleteErr:   nil,
			wantErrDiag: false,
		},
		{
			name: "errors on other conflicts (e.g. in use by derived column)",
			deleteErr: client.DetailedError{
				Status:  http.StatusConflict,
				Message: "Column is in use by 1 derived columns: 'my_dc'",
			},
			wantErrDiag: true,
		},
		{
			name: "errors on non-conflict API errors",
			deleteErr: client.DetailedError{
				Status:  http.StatusInternalServerError,
				Message: "internal server error",
			},
			wantErrDiag: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &columnResource{
				client: &client.Client{
					Columns: mockColumns{deleteErr: tt.deleteErr},
				},
			}
			var resp tfresource.DeleteResponse
			r.Delete(ctx, tfresource.DeleteRequest{State: state}, &resp)
			assert.Equal(t, tt.wantErrDiag, resp.Diagnostics.HasError())
		})
	}
}

func testAccEnsureColumnExists(t *testing.T, resource, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("%q not found in state", resource)
		}

		client := testAccClient(t)
		_, err := client.Columns.GetByKeyName(context.Background(), resourceState.Primary.Attributes["dataset"], name)
		if err != nil {
			return fmt.Errorf("failed to fetch created column: %w", err)
		}

		return nil
	}
}
