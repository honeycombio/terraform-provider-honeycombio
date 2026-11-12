package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	hnyclient "github.com/honeycombio/terraform-provider-honeycombio/client"
	v2client "github.com/honeycombio/terraform-provider-honeycombio/client/v2"
)

const testAccTelemetrySchemaA = `groups:
  - id: registry.acceptance
    type: attribute_group
    brief: Attributes emitted by the acceptance tests.
    attributes:
      - id: acceptance.order.id
        type: string
        brief: The ID of the order.
        stability: stable
`

// testAccTelemetrySchemaB is formatting a normalising client would change:
// comments, aligned values, trailing whitespace, blank lines and no final newline.
const testAccTelemetrySchemaB = `# Telemetry schema for the acceptance tests.

groups:
  - id:         registry.acceptance   
    type:       attribute_group
    brief:      "Attributes emitted by the acceptance tests."   # trailing comment


    attributes:
      - id: acceptance.order.id
        type: string
        brief: The ID of the order.
        stability: stable`

// The tests below share the Team's single telemetry schema,
// so they must not run in parallel with each other.

func TestAcc_TelemetrySchemaResource(t *testing.T) {
	ctx := context.Background()
	c := testAccV2Client(t)
	// the Team has one schema, so make sure this test leaves it unset
	t.Cleanup(func() { _ = c.TelemetrySchemas.Delete(ctx) })

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6MuxServerFactory,
		CheckDestroy:             testAccEnsureTelemetrySchemaDeleted(t),
		Steps: []resource.TestStep{
			{ // creating over an existing schema is refused, whatever it holds
				PreConfig: func() {
					_, err := c.TelemetrySchemas.Put(ctx, &v2client.TelemetrySchema{SchemaYAML: testAccTelemetrySchemaA})
					require.NoError(t, err)
				},
				Config:      testAccConfigTelemetrySchema(testAccTelemetrySchemaA),
				ExpectError: regexp.MustCompile(`(?s)Telemetry Schema Already Exists`),
			},
			{ // the refused create left the existing schema alone
				PreConfig: func() {
					require.NoError(t, testAccEnsureTelemetrySchemaIs(t, testAccTelemetrySchemaA)(nil))
					require.NoError(t, c.TelemetrySchemas.Delete(ctx))
				},
				Config: testAccConfigTelemetrySchema(testAccTelemetrySchemaA),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("honeycombio_telemetry_schema.test", "id"),
					resource.TestCheckResourceAttr("honeycombio_telemetry_schema.test", "schema_yaml", testAccTelemetrySchemaA),
					testAccEnsureTelemetrySchemaIs(t, testAccTelemetrySchemaA),
				),
			},
			{ // update to a document with unusual formatting, which must come back untouched
				Config: testAccConfigTelemetrySchema(testAccTelemetrySchemaB),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("honeycombio_telemetry_schema.test", "schema_yaml", testAccTelemetrySchemaB),
					testAccEnsureTelemetrySchemaIs(t, testAccTelemetrySchemaB),
				),
			},
			{ // import by the Team's slug gives the same state
				ResourceName:      "honeycombio_telemetry_schema.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // another Team's slug is not importable with this key
				ResourceName:  "honeycombio_telemetry_schema.test",
				ImportState:   true,
				ImportStateId: "not-the-slug-of-this-team",
				ExpectError:   regexp.MustCompile(`Cannot import non-existent remote object`),
			},
			{ // a change made outside Terraform is detected
				PreConfig: func() {
					_, err := c.TelemetrySchemas.Put(ctx, &v2client.TelemetrySchema{SchemaYAML: testAccTelemetrySchemaA})
					require.NoError(t, err)
				},
				Config:             testAccConfigTelemetrySchema(testAccTelemetrySchemaB),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{ // and put right again by an apply
				Config: testAccConfigTelemetrySchema(testAccTelemetrySchemaB),
				Check:  testAccEnsureTelemetrySchemaIs(t, testAccTelemetrySchemaB),
			},
			{ // a schema deleted outside Terraform is created again
				PreConfig: func() {
					require.NoError(t, c.TelemetrySchemas.Delete(ctx))
				},
				Config: testAccConfigTelemetrySchema(testAccTelemetrySchemaB),
				Check:  testAccEnsureTelemetrySchemaIs(t, testAccTelemetrySchemaB),
			},
		},
	})
}

// TestAcc_TelemetrySchemaResource_adopt follows the way to bring a schema which
// already exists under management: import it, then plan.
func TestAcc_TelemetrySchemaResource_adopt(t *testing.T) {
	ctx := context.Background()
	c := testAccV2Client(t)
	t.Cleanup(func() { _ = c.TelemetrySchemas.Delete(ctx) })

	auth, err := c.AuthInfo(ctx)
	require.NoError(t, err)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6MuxServerFactory,
		CheckDestroy:             testAccEnsureTelemetrySchemaDeleted(t),
		Steps: []resource.TestStep{
			{ // import a schema which was set outside Terraform
				PreConfig: func() {
					_, err := c.TelemetrySchemas.Put(ctx, &v2client.TelemetrySchema{SchemaYAML: testAccTelemetrySchemaB})
					require.NoError(t, err)
				},
				Config:             testAccConfigTelemetrySchema(testAccTelemetrySchemaB),
				ResourceName:       "honeycombio_telemetry_schema.test",
				ImportState:        true,
				ImportStateId:      auth.Team.Slug,
				ImportStatePersist: true,
			},
			{ // the configuration matches, so there is nothing to change
				Config:   testAccConfigTelemetrySchema(testAccTelemetrySchemaB),
				PlanOnly: true,
			},
			{ // a configuration which differs shows the change before it replaces the schema
				Config:             testAccConfigTelemetrySchema(testAccTelemetrySchemaA),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAcc_TelemetrySchemaResource_validation covers the checks made before the
// API is called, so it passes even where the telemetry schema API is unavailable.
func TestAcc_TelemetrySchemaResource_validation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheckV2API(t),
		ProtoV6ProviderFactories: testAccProtoV6MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config:      testAccConfigTelemetrySchema(""),
				ExpectError: regexp.MustCompile(`string length must be between 1 and 1048576`),
			},
			{
				// 600,000 characters but 1,200,000 bytes
				Config:      testAccConfigTelemetrySchema(strings.Repeat("é", 600_000)),
				ExpectError: regexp.MustCompile(`string length must be between 1 and 1048576`),
			},
		},
	})
}

// testAccConfigTelemetrySchema returns a configuration for the resource
// holding schema exactly, using a JSON string which is also a valid HCL string.
func testAccConfigTelemetrySchema(schema string) string {
	quoted, err := json.Marshal(schema)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf(`
resource "honeycombio_telemetry_schema" "test" {
  schema_yaml = %s
}`, quoted)
}

// testAccEnsureTelemetrySchemaIs checks the API returns schema, byte for byte.
func testAccEnsureTelemetrySchemaIs(t *testing.T, schema string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		got, err := testAccV2Client(t).TelemetrySchemas.Get(context.Background())
		if err != nil {
			return fmt.Errorf("failed to fetch the telemetry schema: %w", err)
		}
		if got.SchemaYAML != schema {
			return fmt.Errorf("the API returned a different schema:\nwant: %q\n got: %q", schema, got.SchemaYAML)
		}
		return nil
	}
}

func testAccEnsureTelemetrySchemaDeleted(t *testing.T) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		_, err := testAccV2Client(t).TelemetrySchemas.Get(context.Background())
		var de hnyclient.DetailedError
		if errors.As(err, &de) && de.IsNotFound() {
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to check the telemetry schema was deleted: %w", err)
		}
		return errors.New("the telemetry schema still exists")
	}
}
