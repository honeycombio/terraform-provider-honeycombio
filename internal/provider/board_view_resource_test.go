package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
