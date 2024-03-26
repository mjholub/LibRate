package searchdb

type targetDBName string

const (
	Members           targetDBName = "members"
	Artists           targetDBName = "artists"
	Ratings           targetDBName = "ratings"
	Genres            targetDBName = "genres"
	GenreDescriptions targetDBName = "genre_descriptions"
	// aka keywords
	Studios targetDBName = "studio"
	MediaDB targetDBName = "media"
)

// nolint:gochecknoglobals
var AllTargets = []TargetDB{
	Members, Artists, Ratings, Genres,
	GenreDescriptions, Studios, MediaDB,
}

type TargetDB interface {
	String() string
}

func (t targetDBName) String() string {
	return string(t)
}
