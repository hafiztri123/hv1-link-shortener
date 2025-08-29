package url

import (
	"math"
	"strings"
)

const (
	base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	base        = uint64(len(base62Chars))
)

func toBase62(urlId uint64) string {
	if urlId == 0 {
		return string(base62Chars[0])
	}

	var sb strings.Builder
	for urlId > 0 {
		remainder := urlId % base
		sb.WriteByte(base62Chars[remainder])
		urlId /= base
	}

	return reverse(sb.String())
}

func fromBase62(shortCode string) uint64 {
	var n uint64
	for i, char := range shortCode {
		power := len(shortCode) - (i + 1)
		pos := strings.IndexRune(base62Chars, char)
		n += uint64(pos) * uint64(math.Pow(float64(base), float64(power)))
	}
	return n
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
