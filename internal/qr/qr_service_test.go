package qr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestGenerateQRCode(t *testing.T) {
	testCases := []struct {
		name string
		url string
		wantErr bool
	}{
		{
			name: "success",
			url: "https://example.com",
			wantErr: false,
		},
		{
			name: "url with no http",
			url: "example.com",
			wantErr: false,
		},
		{
			name: "invalid url",
			url: "",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := GenerateQRCode(tc.url)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			
		})
	}
}