package main

import "os"

func main() {
	parseFlags()
	err := run()
	if err != nil {
		os.Exit(1)
	}
}
