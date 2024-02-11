package searchdb

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

// nolint:gochecknoglobals
var AllTargets = []TargetDB{
	Members, ArtistsGroup, ArtistsIndividual, Ratings, Genres,
	GenreDescriptions, GenreCharacteristics, Studios,
}

type TargetDB interface {
	String() string
}

func (t targetDBName) String() string {
	return string(t)
}
