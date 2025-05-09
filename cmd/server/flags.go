package main

import (
	"flag"
	"os"
)

var (
	flagRunAddress string
	flagLogLevel   string
)

func parseFlags() {
	flag.StringVar(&flagRunAddress, "a", ":8080", "Address and port to run server")
	flag.StringVar(&flagLogLevel, "l", "info", "Log level (e.g., debug, info, warn, error)")

	flag.Parse()

	if envRunAddress := os.Getenv("ADDRESS"); envRunAddress != "" {
		flagRunAddress = envRunAddress
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		flagLogLevel = envLogLevel
	}

}
