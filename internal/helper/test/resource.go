package test

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestCheckResourceAttrAtLeast returns a resource.TestCheckFunc that checks that the
// given attribute is at least the given value.
//
// This is useful for checking that a bunch of resources has been created but not caring
// about the *exact* number.
func TestCheckResourceAttrAtLeast(name, key string, value int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		raw, ok := rs.Primary.Attributes[key]
		if !ok {
			return fmt.Errorf("attribute not found: %s", key)
		}

		v, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("failed to parse attribute %s: %w", key, err)
		}

		if v < value {
			return fmt.Errorf("expected %s to be at least %d, got %d", key, value, v)
		}

		return nil
	}
}
