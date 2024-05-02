package honeycombio

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// testCheckOutputContains checks an output in the Terraform configuration
// contains the given value. The output is expected to be of type list(string).
func testCheckOutputContains(name, contains string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Outputs[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		output := rs.Value.([]interface{})

		for _, value := range output {
			if value.(string) == contains {
				return nil
			}
		}

		return fmt.Errorf("Output '%s' did not contain %#v, got %#v", name, contains, output)
	}
}

// testCheckOutputDoesNotContain checks an output in the Terraform configuration
// does not contain the given value. The output is expected to be of type
// list(string).
func testCheckOutputDoesNotContain(name, contains string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Outputs[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		output := rs.Value.([]interface{})

		for _, value := range output {
			if value.(string) == contains {
				return fmt.Errorf("Output '%s' contained %#v, should not", name, contains)
			}
		}

		return nil
	}
}

// MinifyJSON minifies a JSON string removing all whitespace and newlines
func MinifyJSON(s string) (string, error) {
	var buffer bytes.Buffer
	err := json.Compact(&buffer, []byte(s))
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}
