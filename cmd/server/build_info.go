package main

import "fmt"

// buildVersion содержит версию сборки.
// Значение устанавливается через флаг -ldflags на этапе компиляции.
var buildVersion string

// buildDate содержит дату сборки.
// Значение устанавливается через флаг -ldflags на этапе компиляции.
var buildDate string

// buildCommit содержит хеш коммита из системы контроля версий.
// Значение устанавливается через флаг -ldflags на этапе компиляции.
var buildCommit string

// printBuildInfo выводит информацию о версии сборки.
// Если значения не заданы, выводит "N/A".
func printBuildInfo() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
