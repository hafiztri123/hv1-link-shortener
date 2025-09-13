package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlaceholders(t *testing.T) {
	placeholder := PlaceholderBuilder(5, 1)
	assert.Equal(t, "$1,$2,$3,$4,$5", placeholder)
}
