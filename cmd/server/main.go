package main

func main() {
	printBuildInfo()

	err := parseFlags()
	if err != nil {
		panic(err)
	}

	err = run()
	if err != nil {
		panic(err)
	}
}
