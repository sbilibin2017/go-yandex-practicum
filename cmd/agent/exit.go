package main

func exit(err error) int {
	if err != nil {
		return 1
	}
	return 0
}
