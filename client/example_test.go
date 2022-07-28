package client

import (
	"context"
	"fmt"
	"os"
)

// This example shows how to configure and use a client to list all boards.
func Example() {
	apiKey, _ := os.LookupEnv("HONEYCOMB_API_KEY")

	config := &Config{
		APIKey: apiKey,
	}
	client, err := NewClient(config)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	boards, err := client.Boards.List(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d boards\n", len(boards))

	for i, board := range boards {
		fmt.Printf("%d| %s (%d queries)\n", i, board.Name, len(board.Queries))
	}
}
