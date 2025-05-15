package main

import (
	"context"
	"os"
)

func main() {
	opts := parseFlags()
	err := run(context.Background(), opts)
	if err != nil {
		os.Exit(1)
	}
}
