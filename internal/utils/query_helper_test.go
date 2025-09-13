package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlaceholders(t *testing.T) {
	placeholder := SelectPlaceholderBuilder(5, 1)
	assert.Equal(t, "$1,$2,$3,$4,$5", placeholder)
}

func TestPlaceholdersEdgeCases(t *testing.T) {
	testCases := []struct {
		name      string
		dataCount int
		offset    int
		expected  string
	}{
		{
			name:      "zero count returns empty string",
			dataCount: 0,
			offset:    1,
			expected:  "",
		},
		{
			name:      "single placeholder with offset 1",
			dataCount: 1,
			offset:    1,
			expected:  "$1",
		},
		{
			name:      "multiple placeholders with offset 0",
			dataCount: 3,
			offset:    0,
			expected:  "$0,$1,$2",
		},
		{
			name:      "multiple placeholders with different offset",
			dataCount: 3,
			offset:    5,
			expected:  "$5,$6,$7",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SelectPlaceholderBuilder(tc.dataCount, tc.offset)
			assert.Equal(t, tc.expected, result)
		})
	}
}
