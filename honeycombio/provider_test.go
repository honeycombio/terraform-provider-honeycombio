package honeycombio

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/joho/godotenv"
)

func init() {
	// load environment values from a .env, if available
	_ = godotenv.Load("../.env")
}

func testAccPreCheck(t *testing.T) func() {
	return func() {
		if _, ok := os.LookupEnv("HONEYCOMBIO_APIKEY"); !ok {
			t.Fatalf("environment variable HONEYCOMBIO_APIKEY must be set to run acceptance tests")
		}
		if _, ok := os.LookupEnv("HONEYCOMBIO_DATASET"); !ok {
			t.Fatalf("environment variable HONEYCOMBIO_DATASET must be set to run acceptance tests")
		}
	}
}

var testAccProviderFactories = map[string]func() (*schema.Provider, error){
	"honeycombio": func() (*schema.Provider, error) {
		return Provider(), nil
	},
}

func testAccClient(t *testing.T) *honeycombio.Client {
	cfg := &honeycombio.Config{
		APIKey: os.Getenv("HONEYCOMBIO_APIKEY"),
	}
	c, err := honeycombio.NewClient(cfg)
	if err != nil {
		t.Fatalf("could not initialize honeycombio.Client: %v", err)
	}
	return c
}

func testAccDataset() string {
	return os.Getenv("HONEYCOMBIO_DATASET")
}
