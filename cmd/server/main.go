package main

import "context"

func main() {
	parseFlags()
	err := run(context.Background())
	exit(err)
}
