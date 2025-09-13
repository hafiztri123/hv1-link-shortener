package utils

import (
	"fmt"
	"strings"
)

func SelectPlaceholderBuilder(dataCount int, offset int) string {
	if dataCount == 0 {
		return ""
	}

	placeholders := make([]string, dataCount)
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+offset)
	}

	return strings.Join(placeholders, ",")
}
