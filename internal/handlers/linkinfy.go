package handlers

import (
	"strings"

	lop "github.com/samber/lo/parallel"
)

// say we got an element "Tribal Ambient"
// this must become "https://$LIBRATE_HOST/genres/music/tribal-ambient"
// c.BaseUrl() returns "https://$LIBRATE_HOST"
// we can further combine this with the "genres/music/" in caller
// but we need an anchor element, so
// <a href="https://$LIBRATE_HOST/genres/music/tribal-ambient">Tribal Ambient</a>

func LinksFromArray(prefix string, arr []string, suffix ...string) []string {
	return lop.Map(arr, func(base string, _ int) string {
		baseURLFormat := strings.ToLower(strings.ReplaceAll(base, " ", "-"))
		if len(suffix) > 0 {
			return "<a href=\"" + prefix + baseURLFormat + strings.Join(suffix, "") + "\">" + base + "</a>"
		}
		return "<a href=\"" + prefix + baseURLFormat + "\">" + base + "</a>"
	})
}
