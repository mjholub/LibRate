package aggregation

type (
	mediaAggregation  string
	userAggregation   string
	postAggregation   string
	artistAggregation string
)

const (
	// RatingCount does not take reviews with a text body into account
	RatingCount mediaAggregation = "rating_count"
	// Contentious is a special aggregation, that is used to determine the most
	// polarizing media, ones that people either love or hate.
	Contentious mediaAggregation = "contentious"
	// ReviewCount is useful when one wants to know more
	// about a media than just it's core and synopsis (if any)
	ReviewCount mediaAggregation = "review_count"
	// RewardCount counts the number of received rewards
	RewardCount mediaAggregation = "reward_count"
	// NominationsCount counts the number of nominations to awards
	NominationsCount mediaAggregation = "nominations_count"
	// AverageRating is the average rating of a media item
	// Reference scale is 0-1000 (since integer operations are more efficient),
	// which by default is then displayed as 0-100 float with one decimal,
	// but users can set their own scale relative to that.
	AverageRating mediaAggregation = "average_rating"
	// WeightedScore is the average rating of a media item
	// weighed by the number of ratings
	WeightedScore mediaAggregation = "weighted_score"
	// Added is the date when the media was added to the database
	Added mediaAggregation = "added"
	// Modified is the date when the media was last modified
	Modified mediaAggregation = "modified"
	// MostScrobled is the number of times a media item has been scrobbled
	MostScrobbled mediaAggregation = "most_scrobbled"
	// TopCountries returns the countries where the media is most popular
	TopCountries mediaAggregation = "top_countries"
)

const (
	// MemberSince is the date when the member joined the platform
	MemberSince userAggregation = "member_since"
	// LastActive is the date when the member was last active
	LastActive userAggregation = "last_active"
	// PostCount is the number of indexable posts a user has made, including reviews
	PostCount userAggregation = "post_count"
	// WrittenReviews is the number of reviews a user has made
	// Can't be declared as "ReviewsCount" due to NS conflict with media
	WrittenReviews userAggregation = "review_count"
	// NOTE: I don't really want to create another
	// reddit clone with it's karma cancer, so unsure about this
	// ReviewVotes is the upvote score of a user's reviews
	ReviewVotes userAggregation = "review_votes"
	// TopMusicGenres is the list of music genres
	// for which the user has given the highest average rating
	TopMusicGenres userAggregation = "top_music_genres"
	TopFilmGenres  userAggregation = "top_film_genres"
	TopGameGenres  userAggregation = "top_game_genres"
	TopBookGenres  userAggregation = "top_book_genres"
	BooksRead      userAggregation = "books_read"
	FilmsWatched   userAggregation = "films_watched"
	GamesPlayed    userAggregation = "games_played"
	AlbumsListened userAggregation = "albums_listened"
)

const (
	// ReactionCount is the number of reactions a post has received
	ReactionsCount postAggregation = "reactions_count"
	// RepliesCount is the number of replies a post has received
	RepliesCount postAggregation = "replies_count"
	// IsReview is a boolean aggregation, that is used to determine
	// whether the post is a review
	IsReview postAggregation = "is_review"
	// AuthorSeparationDegree is the number of degrees of separation
	// between the author of the post and the current user
	AuthorSeparationDegree postAggregation = "author_separation_degree"
	// PostDate is the date when the post was made
	PostDate postAggregation = "post_date"
)

const (
	// WorksCount is the number of works an artist has made
	// or if it's a label/studio/publishing house etc, released
	// It is non-discriminatory, so for example it can list
	// films where an actor hasn't been included in the credits
	WorksCount artistAggregation = "works_count"
	// MainWorksCount is the number of works an artist has made
	// or played a significant role in making
	MainWorksCount artistAggregation = "main_works_count"
	// ArtistType is the type of the artist.
	// Can be 'individual', 'group', 'studio'
	ArtistType artistAggregation = "artist_type"
	// Active is a boolean aggregation, that is used to determine
	// whether the artist is active
	Active artistAggregation = "active"
	// AverageWorksRating is the average rating of an artist's works
	AverageWorksRating artistAggregation = "average_rating"
	// AverageWorksRatingByDate is the average rating of an artist's works
	// by date range
	AverageWorksRatingByDate artistAggregation = "average_rating_by_date"
	// TopGenres is the list of most common genres of the artist's works
	TopGenres artistAggregation = "top_genres"
	// AssociatedActs is the list of projects the artist has been associated with
	AssociatedActs artistAggregation = "associated_acts"
	// Location is the location of the artist
	// It applies to both the birthplace and the current residence
	Location artistAggregation = "location"
	// ReviewsReceived is the number of reviews the artist has received
	ReviewsReceived artistAggregation = "reviews_received"
	// RewardsReceived is the number of rewards the artist has received
	RewardsReceived artistAggregation = "rewards_received"
)

type AnyAggregation interface {
	artistAggregation | postAggregation | userAggregation | mediaAggregation
}

type Aggregation interface {
	stringValue() string
}

func (s mediaAggregation) stringValue() string {
	return string(s)
}

func (s userAggregation) stringValue() string {
	return string(s)
}

func (s postAggregation) stringValue() string {
	return string(s)
}

func (s artistAggregation) stringValue() string {
	return string(s)
}
