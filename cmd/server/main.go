package main

func main() {
	printBuildInfo()
	parseFlags()
	err := run()
	if err != nil {
		panic(err)
	}
}
