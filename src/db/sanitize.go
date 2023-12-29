package db

import (
	"regexp"
)

func Sanitize(s []string) []string {
	// only A-Z, a-z, 0-9, and _ are allowed
	whitelist := regexp.MustCompile(`[^A-Za-z0-9_]`)
	for i := range s {
		s[i] = whitelist.ReplaceAllString(s[i], "")
	}
	return s
}
