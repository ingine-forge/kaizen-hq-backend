package utils

import "strings"

func BuildCustomID(parts ...string) string {
	return strings.Join(parts, "_")
}
