package utils

func StringSliceToAny(strings []string) []any {
	newArray := make([]any, len(strings))
	for i, string := range strings {
		newArray[i] = string
	}
	return newArray
}
