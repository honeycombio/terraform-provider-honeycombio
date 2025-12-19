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
)

func TestAccHoneycombioBoardView(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)

	// Create a board first since views belong to boards
	board, err := c.Boards.Create(ctx, &client.Board{
		Name:      "Test Board " + acctest.RandString(8),
		BoardType: "flexible",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		c.Boards.Delete(ctx, board.ID)
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: testBoardViewConfigBasic(board.ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardViewExists(t, "honeycombio_board_view.test", board.ID),
					resource.TestCheckResourceAttr("honeycombio_board_view.test", "name", "Test Board View"),
					resource.TestCheckResourceAttr("honeycombio_board_view.test", "board_id", board.ID),
					resource.TestCheckResourceAttr("honeycombio_board_view.test", "filter.#", "2"),
					resource.TestCheckResourceAttr("honeycombio_board_view.test", "filter.0.column", "service.name"),
					resource.TestCheckResourceAttr("honeycombio_board_view.test", "filter.0.operation", "exists"),
					resource.TestCheckResourceAttr("honeycombio_board_view.test", "filter.1.column", "duration_ms"),
					resource.TestCheckResourceAttr("honeycombio_board_view.test", "filter.1.operation", ">"),
				),
			},
			{
				Config: testBoardViewConfigWithArrayFilter(board.ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardViewExists(t, "honeycombio_board_view.test", board.ID),
					resource.TestCheckResourceAttr("honeycombio_board_view.test", "name", "Test Board View Updated"),
					resource.TestCheckResourceAttr("honeycombio_board_view.test", "filter.#", "1"),
					resource.TestCheckResourceAttr("honeycombio_board_view.test", "filter.0.column", "environment"),
					resource.TestCheckResourceAttr("honeycombio_board_view.test", "filter.0.operation", "in"),
				),
			},
			{
				ResourceName:            "honeycombio_board_view.test",
				ImportState:             true,
				ImportStateIdFunc:       testAccBoardViewImportStateIdFunc("honeycombio_board_view.test", board.ID),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testBoardViewConfigBasic(boardID string) string {
	return fmt.Sprintf(`
resource "honeycombio_board_view" "test" {
  board_id = "%s"
  name     = "Test Board View"

  filter {
    column    = "service.name"
    operation = "exists"
  }

  filter {
    column    = "duration_ms"
    operation = ">"
    value     = "100"
  }
}
`, boardID)
}

func testBoardViewConfigWithArrayFilter(boardID string) string {
	return fmt.Sprintf(`
resource "honeycombio_board_view" "test" {
  board_id = "%s"
  name     = "Test Board View Updated"

  filter {
    column    = "environment"
    operation = "in"
    value     = "production,staging,development"
  }
}
`, boardID)
}

func testAccCheckBoardViewExists(t *testing.T, name, boardID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		viewID := resourceState.Primary.ID
		if viewID == "" {
			return fmt.Errorf("board view ID is empty")
		}

		client := testAccClient(t)
		_, err := client.BoardViews.Get(context.Background(), boardID, viewID)
		if err != nil {
			return fmt.Errorf("could not find created board view: %w", err)
		}
		return nil
	}
}

func testAccBoardViewImportStateIdFunc(resourceName, boardID string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		viewID := rs.Primary.ID
		if viewID == "" {
			return "", fmt.Errorf("board view ID is empty")
		}

		return fmt.Sprintf("%s/%s", boardID, viewID), nil
	}
}

func TestAccHoneycombioBoardView_validation(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)

	// Create a board first since views belong to boards
	board, err := c.Boards.Create(ctx, &client.Board{
		Name:      "Test Board " + acctest.RandString(8),
		BoardType: "flexible",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		c.Boards.Delete(ctx, board.ID)
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config:      testBoardViewConfigNoFilters(board.ID),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?i)block filter must have a configuration value`),
			},
			{
				// Create a valid board view
				Config: testBoardViewConfigBasic(board.ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardViewExists(t, "honeycombio_board_view.test", board.ID),
					resource.TestCheckResourceAttr("honeycombio_board_view.test", "filter.#", "2"),
				),
			},
			{
				// Try to update to remove all filters - this will fail at schema validation
				// because the block is required
				Config:      testBoardViewConfigNoFilters(board.ID),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?i)block filter must have a configuration value`),
			},
			{
				// Final step with valid configuration to allow cleanup
				Config: testBoardViewConfigBasic(board.ID),
			},
		},
	})
}

func TestAccHoneycombioBoardView_emptyStringInArrayFilter(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)

	// Create a board first since views belong to boards
	board, err := c.Boards.Create(ctx, &client.Board{
		Name:      "Test Board " + acctest.RandString(8),
		BoardType: "flexible",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		c.Boards.Delete(ctx, board.ID)
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				// Test with empty strings in comma-separated list - should fail validation
				Config:      testBoardViewConfigWithEmptyStrings(board.ID),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?i)empty.*comma-separated`),
			},
			{
				// Test with only empty strings - should fail validation
				Config:      testBoardViewConfigWithOnlyEmptyStrings(board.ID),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?i)empty.*comma-separated`),
			},
			{
				// Test with valid comma-separated values (no empty strings)
				Config: testBoardViewConfigWithValidArrayFilter(board.ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardViewExists(t, "honeycombio_board_view.test", board.ID),
					resource.TestCheckResourceAttr("honeycombio_board_view.test", "filter.#", "1"),
					resource.TestCheckResourceAttr("honeycombio_board_view.test", "filter.0.operation", "in"),
					resource.TestCheckResourceAttr("honeycombio_board_view.test", "filter.0.value", "value1,value2,value3"),
				),
			},
		},
	})
}

func testBoardViewConfigWithEmptyStrings(boardID string) string {
	return fmt.Sprintf(`
resource "honeycombio_board_view" "test" {
  board_id = "%s"
  name     = "Test View With Empty Strings"

  filter {
    column    = "environment"
    operation = "in"
    value     = "value1,,value2, ,value3"
  }
}
`, boardID)
}

func testBoardViewConfigWithOnlyEmptyStrings(boardID string) string {
	return fmt.Sprintf(`
resource "honeycombio_board_view" "test" {
  board_id = "%s"
  name     = "Test View With Only Empty Strings"

  filter {
    column    = "environment"
    operation = "in"
    value     = ", ,,"
  }
}
`, boardID)
}

func testBoardViewConfigWithValidArrayFilter(boardID string) string {
	return fmt.Sprintf(`
resource "honeycombio_board_view" "test" {
  board_id = "%s"
  name     = "Test View With Valid Array Filter"

  filter {
    column    = "environment"
    operation = "in"
    value     = "value1,value2,value3"
  }
}
`, boardID)
}

func testBoardViewConfigNoFilters(boardID string) string {
	return fmt.Sprintf(`
resource "honeycombio_board_view" "test" {
  board_id = "%s"
  name     = "Test Board View Without Filters"
}
`, boardID)
}

func TestAccHoneycombioBoardView_boardIdRequiresReplace(t *testing.T) {
	ctx := context.Background()
	c := testAccClient(t)

	// Create two boards
	board1, err := c.Boards.Create(ctx, &client.Board{
		Name:      "Test Board 1 " + acctest.RandString(8),
		BoardType: "flexible",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		c.Boards.Delete(ctx, board1.ID)
	})

	board2, err := c.Boards.Create(ctx, &client.Board{
		Name:      "Test Board 2 " + acctest.RandString(8),
		BoardType: "flexible",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		c.Boards.Delete(ctx, board2.ID)
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				// Create board view on board1
				Config: testBoardViewConfigForBoard(board1.ID, "Test View"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardViewExists(t, "honeycombio_board_view.test", board1.ID),
					resource.TestCheckResourceAttr("honeycombio_board_view.test", "board_id", board1.ID),
				),
			},
			{
				// Change board_id to board2 - this should trigger a replacement
				Config: testBoardViewConfigForBoard(board2.ID, "Test View"),
				// Verify the plan shows replacement
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("honeycombio_board_view.test", plancheck.ResourceActionReplace),
					},
				},
			},
			{
				// Apply the change and verify it works
				Config: testBoardViewConfigForBoard(board2.ID, "Test View"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardViewExists(t, "honeycombio_board_view.test", board2.ID),
					resource.TestCheckResourceAttr("honeycombio_board_view.test", "board_id", board2.ID),
				),
			},
		},
	})
}

func testBoardViewConfigForBoard(boardID, name string) string {
	return fmt.Sprintf(`
resource "honeycombio_board_view" "test" {
  board_id = "%s"
  name     = "%s"

  filter {
    column    = "service.name"
    operation = "exists"
  }
}
`, boardID, name)
}
