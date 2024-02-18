package search

import (
	"slices"

	"codeberg.org/mjh/LibRate/controllers/search/target"
)

func filterByCategories(category target.Category) []string {
	reviewFields := []string{
		"media",
		"topic",
		"comment",
		"date",
	}
	artistsFields := []string{
		"artist_name",
		"roles",
		"country",
		"bio",
		"added",
		"modified",
		"active",
	}
	usersFields := []string{
		"webfinger",
		"display_name",
		"bio",
	}

	mediaFields := []string{
		"kind",
		"title",
		"artists",
		"genres",
		"language",
		"released",
		"added",
		"modified",
	}

	genreFields := []string{
		"name",
		"kinds",
		"descriptions",
		"language",
	}

	switch category {
	case target.Artists:
		return artistsFields
	case target.Media:
		return mediaFields
	case target.Users:
		return usersFields
	case target.Reviews, target.Posts:
		return reviewFields
	case target.Genres:
		return genreFields
	default:
		return slices.Concat(reviewFields, artistsFields, mediaFields, usersFields)
	}
}
