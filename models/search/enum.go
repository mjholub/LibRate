package searchdb

import "fmt"

type targetDBName string

const (
	Members           targetDBName = "members"
	ArtistsIndividual targetDBName = "person"
	ArtistsGroup      targetDBName = "group"
	Ratings           targetDBName = "ratings"
	Genres            targetDBName = "genres"
	GenreDescriptions targetDBName = "genre_descriptions"
	// aka keywords
	GenreCharacteristics targetDBName = "genre_characteristics"
	Studios              targetDBName = "studio"
)

type TargetDB interface {
	String() string
}

func (t targetDBName) String() string {
	return fmt.Sprint(t)
}
