package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestStringsSliceToAny(t *testing.T) {
	input := []string{"0", "1", "2"}
	expected := []any{"0", "1", "2"}

	actual := StringSliceToAny(input)


	assert.Equal(t,expected, actual) 
}
