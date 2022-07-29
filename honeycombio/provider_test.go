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
		if _, ok := os.LookupEnv("HONEYCOMB_API_KEY"); !ok {
			t.Fatalf("environment variable HONEYCOMB_API_KEY must be set to run acceptance tests")
		}
		if _, ok := os.LookupEnv("HONEYCOMB_DATASET"); !ok {
			t.Fatalf("environment variable HONEYCOMB_DATASET must be set to run acceptance tests")
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
		APIKey: os.Getenv("HONEYCOMB_API_KEY"),
	}
	c, err := honeycombio.NewClient(cfg)
	if err != nil {
		t.Fatalf("could not initialize honeycombio.Client: %v", err)
	}
	return c
}

func testAccDataset() string {
	return os.Getenv("HONEYCOMB_DATASET")
}

func testAccDatasetWithArgs(createArgs honeycombio.DatasetCreateArgs) honeycombio.Dataset {
	dataset := honeycombio.Dataset{
		Name:            createArgs.Name,
		Description:     &createArgs.Description,
		ExpandJSONDepth: &createArgs.ExpandJSONDepth,
	}
	return dataset
}
