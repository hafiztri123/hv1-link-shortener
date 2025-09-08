package qr

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/skip2/go-qrcode"
)

func GenerateQRCode(url string) ([]byte, error) {

	if url == "" {
		return nil, errors.New("URL is empty")
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	qrBytes, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		slog.Error("Failed to generate qr code", "error", err)
		return nil, err
	}

	return qrBytes, nil
}
