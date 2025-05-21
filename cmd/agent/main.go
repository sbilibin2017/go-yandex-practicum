package main

func main() {
	parseFlags()
	err := run()
	if err != nil {
		panic(err)
	}
}
