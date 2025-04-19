package utilities

import "strings"

func NormalizeTag(tag string) string {
	return strings.TrimSpace(strings.ToLower(tag))
}
