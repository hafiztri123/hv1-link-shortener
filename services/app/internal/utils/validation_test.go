package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURLValidation(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid url",
			input: "https://example.com",
			want:  true,
		},
		{
			name:  "wrong scheme",
			input: "javascript://example.com",
			want:  false,
		},
		{
			name:  "cant be parsed",
			input: "example.com",
			want:  false,
		},
		{
			name:  "missing host",
			input: "https://",
			want:  false,
		},
		{
			name:  "loopback address",
			input: "http://127.0.0.1",
			want:  false,
		},
		{
			name:  "invalid domain",
			input: "https://this-hostname-definitely-does-not-exist.invalid",
			want:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ok := IsValidURL(tc.input)
			assert.Equal(t, tc.want, ok)
		})
	}

}
