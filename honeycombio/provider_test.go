package honeycombio

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var testAccProvider *schema.Provider
var testAccProviders map[string]terraform.ResourceProvider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]terraform.ResourceProvider{
		"honeycombio": testAccProvider,
	}
}

func testAccPreCheck(t *testing.T) {
	if _, ok := os.LookupEnv("HONEYCOMBIO_APIKEY"); !ok {
		t.Fatalf("environment variable HONEYCOMBIO_APIKEY must be set to run acceptance tests")
	}
	if _, ok := os.LookupEnv("HONEYCOMBIO_DATASET"); !ok {
		t.Fatalf("environment variable HONEYCOMBIO_DATASET must be set to run acceptance tests")
	}
}

func testAccDataset() string {
	return os.Getenv("HONEYCOMBIO_DATASET")
}
