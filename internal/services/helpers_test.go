package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	FieldOne string
	FieldTwo int
}

func TestStructToMap(t *testing.T) {
	testStruct := TestStruct{
		FieldOne: "Value1",
		FieldTwo: 42,
	}
	result := structToMap(testStruct)
	assert.Equal(t, "Value1", result["fieldone"], "fieldone должно быть равно Value1")
	assert.Equal(t, 42, result["fieldtwo"], "fieldtwo должно быть равно 42")
}
