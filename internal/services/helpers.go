package services

import (
	"strings"

	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
)

func structToMap[T any](v T) map[string]any {
	m := structs.Map(v)
	for k, v := range m {
		delete(m, k)
		m[strings.ToLower(k)] = v
	}
	return m
}

func mapToStruct[T any](data map[string]any) (T, error) {
	var result T
	config := &mapstructure.DecoderConfig{
		TagName:     "mapstructure",
		Result:      &result,
		ErrorUnused: true,
	}
	decoder, _ := mapstructure.NewDecoder(config)
	err := decoder.Decode(data)
	if err != nil {
		return result, err
	}
	return result, nil
}

func mapSliceToStructSlice[T any](data []map[string]any) ([]T, error) {
	var result []T
	for _, dataItem := range data {
		item, err := mapToStruct[T](dataItem)
		if err != nil {
			return []T{}, err
		}
		result = append(result, item)
	}
	if result == nil {
		result = []T{}
	}
	return result, nil
}
