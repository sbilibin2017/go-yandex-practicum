package services

import (
	"strings"

	"github.com/fatih/structs"
)

func structToMap(v any) map[string]any {
	m := structs.Map(v)
	for k, v := range m {
		m[strings.ToLower(k)] = v
	}
	return m
}
