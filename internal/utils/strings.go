package utils

import "strings"

func ContainsInsensitive(valA string, valB string) bool {
	return strings.Contains(strings.ToLower(valA), strings.ToLower(valB))
}
