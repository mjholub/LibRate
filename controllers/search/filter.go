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
		"favoriteCount",
		"reblogCount",
	}
	artistsFields := []string{
		"artist_name",
		"roles",
		"country",
		"bio",
		"added",
		"modified",
		"active",
		"associatedArtists",
	}
	usersFields := []string{
		"webfinger",
		"instance",
		"local",
		"displayName",
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
		"description",
		"language",
		"characteristics",
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
