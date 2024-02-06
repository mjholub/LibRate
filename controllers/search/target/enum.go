package target

type category string

const (
	Users   category = "users"
	Artists category = "artists"
	Media   category = "media"
	Groups  category = "groups"
	Tags    category = "tags"
	// posts includes reviews in it's results
	Posts   category = "posts"
	Reviews category = "reviews"
	Genres  category = "genres"
	// aka everything. Default in buildIndexMapping switch
	Union category = "union"
)

type Category interface {
	dummyMethod()
}

func (s category) dummyMethod() {}
