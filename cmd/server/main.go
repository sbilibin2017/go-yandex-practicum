package main

import "context"

func main() {
	err := run(context.Background())
	if err != nil {
		panic(err)
	}
}
