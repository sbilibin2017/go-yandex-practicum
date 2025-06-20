package main

func main() {
	err := flags()
	if err != nil {
		panic(err)
	}
	err = run()
	if err != nil {
		panic(err)
	}
}
