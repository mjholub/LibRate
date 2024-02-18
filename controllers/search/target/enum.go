package target

type category string

const (
	Members category = "members"
	Artists category = "artists"
	Media   category = "media"
	Ratings category = "ratings"
	Studios category = "studios"
	Genres  category = "genres"
	// aka everything. Default in buildIndexMapping switch
	Union category = "union"
)

type Category interface {
	String() string
}

func (s category) String() string {
	return string(s)
}

// we already have a ValidateCategory method,
// no need to return error here
func FromStr(s string) category {
	return category(s)
}
