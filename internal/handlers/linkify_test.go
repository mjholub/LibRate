package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinksFromArray(t *testing.T) {
	tc := []struct {
		name   string
		arr    []string
		prefix string
		suffix []string
		want   []string
	}{
		{
			name:   "music genre: Tribal Ambient",
			arr:    []string{"Tribal Ambient"},
			prefix: "https://librate.app/genres/music/",
			want:   []string{"<a href=\"https://librate.app/genres/music/tribal-ambient\">Tribal Ambient</a>"},
		},
		{
			name:   "multiple music genres",
			arr:    []string{"Dance-Pop", "Blue-Eyed Soul", "Sophisti-Pop", "Italo-Disco", "Adult Contemporary"},
			prefix: "https://librate.app/genres/music/",
			want: []string{
				"<a href=\"https://librate.app/genres/music/dance-pop\">Dance-Pop</a>",
				"<a href=\"https://librate.app/genres/music/blue-eyed-soul\">Blue-Eyed Soul</a>",
				"<a href=\"https://librate.app/genres/music/sophisti-pop\">Sophisti-Pop</a>",
				"<a href=\"https://librate.app/genres/music/italo-disco\">Italo-Disco</a>",
				"<a href=\"https://librate.app/genres/music/adult-contemporary\">Adult Contemporary</a>",
			},
		},
		{
			name:   "unapproved genre query suffix",
			arr:    []string{"Skibidi Music"},
			prefix: "https://librate.app/genres/music/",
			want: []string{
				"<a href=\"https://librate.app/genres/music/skibidi-music?approved=false\">Skibidi Music</a>",
			},
			suffix: []string{"?approved=false"},
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			got := LinksFromArray(c.prefix, c.arr, c.suffix...)
			assert.Equal(t, c.want, got)
		})
	}
}
