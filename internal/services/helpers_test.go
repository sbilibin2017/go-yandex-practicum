package services

import (
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

type Person struct {
	Name string `mapstructure:"name"`
	Age  int    `mapstructure:"age"`
}

func TestStructToMap(t *testing.T) {
	person := Person{Name: "John", Age: 30}
	result := structToMap(person)
	assert.Equal(t, "John", result["name"])
	assert.Equal(t, 30, result["age"])
}

func TestMapToStruct(t *testing.T) {
	data := map[string]any{
		"name": "Jane",
		"age":  25,
	}
	person, _ := mapToStruct[Person](data)
	assert.Equal(t, "Jane", person.Name)
	assert.Equal(t, 25, person.Age)
}

func TestMapToStruct_DecoderInitError(t *testing.T) {
	type Dummy struct{}
	config := &mapstructure.DecoderConfig{
		TagName:     "mapstructure",
		Result:      nil,
		ErrorUnused: true,
	}
	decoder, err := mapstructure.NewDecoder(config)
	assert.Nil(t, decoder)
	assert.Error(t, err)
}

func TestMapSliceToStructSlice(t *testing.T) {
	data := []map[string]any{
		{"name": "John", "age": 30},
		{"name": "Jane", "age": 25},
	}
	people, err := mapSliceToStructSlice[Person](data)
	assert.NoError(t, err)
	assert.Len(t, people, 2)
	assert.Equal(t, "John", people[0].Name)
	assert.Equal(t, 30, people[0].Age)
	assert.Equal(t, "Jane", people[1].Name)
	assert.Equal(t, 25, people[1].Age)
}

func TestEmptyMapSliceToStructSlice(t *testing.T) {
	data := []map[string]any{}
	people, err := mapSliceToStructSlice[Person](data)
	assert.NoError(t, err)
	assert.Len(t, people, 0)
}
