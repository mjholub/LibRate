package aggregation

import (
	"github.com/samber/lo"
)

type (
	MediaAggregation  string
	UserAggregation   string
	PostAggregation   string
	ArtistAggregation string
)

const (
	// RatingCount does not take reviews with a text body into account
	RatingCount MediaAggregation = "rating_count"
	// Contentious is a special aggregation, that is used to determine the most
	// polarizing media, ones that people either love or hate.
	Contentious MediaAggregation = "contentious"
	// ReviewCount is useful when one wants to know more
	// about a media than just it's core and synopsis (if any)
	ReviewCount MediaAggregation = "review_count"
	// RewardCount counts the number of received rewards
	RewardCount MediaAggregation = "reward_count"
	// NominationsCount counts the number of nominations to awards
	NominationsCount MediaAggregation = "nominations_count"
	// AverageRating is the average rating of a media item
	// Reference scale is 0-1000 (since integer operations are more efficient),
	// which by default is then displayed as 0-100 float with one decimal,
	// but users can set their own scale relative to that.
	AverageRating MediaAggregation = "average_rating"
	// WeightedScore is the average rating of a media item
	// weighed by the number of ratings
	WeightedScore MediaAggregation = "weighted_score"
	// Added is the date when the media was added to the database
	Added MediaAggregation = "added"
	// Modified is the date when the media was last modified
	Modified MediaAggregation = "modified"
	// MostScrobled is the number of times a media item has been scrobbled
	MostScrobbled MediaAggregation = "most_scrobbled"
	// TopCountries returns the countries where the media is most popular
	TopCountries MediaAggregation = "top_countries"
)

// nolint: gochecknoglobals
var MediaAggregations = []MediaAggregation{
	RatingCount, Contentious, ReviewCount, RewardCount, NominationsCount,
	AverageRating, WeightedScore, Added, Modified, MostScrobbled, TopCountries,
}

const (
	// MemberSince is the date when the member joined the platform
	MemberSince UserAggregation = "member_since"
	// LastActive is the date when the member was last active
	LastActive UserAggregation = "last_active"
	// PostCount is the number of indexable posts a user has made, including reviews
	PostCount UserAggregation = "post_count"
	// WrittenReviews is the number of reviews a user has made
	// Can't be declared as "ReviewsCount" due to NS conflict with media
	WrittenReviews UserAggregation = "review_count"
	// NOTE: I don't really want to create another
	// reddit clone with it's karma cancer, so unsure about this
	// ReviewVotes is the upvote score of a user's reviews
	ReviewVotes UserAggregation = "review_votes"
	// TopMusicGenres is the list of music genres
	// for which the user has given the highest average rating
	TopMusicGenres UserAggregation = "top_music_genres"
	TopFilmGenres  UserAggregation = "top_film_genres"
	TopGameGenres  UserAggregation = "top_game_genres"
	TopBookGenres  UserAggregation = "top_book_genres"
	BooksRead      UserAggregation = "books_read"
	FilmsWatched   UserAggregation = "films_watched"
	GamesPlayed    UserAggregation = "games_played"
	AlbumsListened UserAggregation = "albums_listened"
)

// nolint: gochecknoglobals
var UserAggregations = []UserAggregation{
	MemberSince, LastActive, PostCount, WrittenReviews, ReviewVotes,
	TopMusicGenres, TopFilmGenres, TopGameGenres, TopBookGenres,
	BooksRead, FilmsWatched, GamesPlayed, AlbumsListened,
}

const (
	// ReactionCount is the number of reactions a post has received
	ReactionsCount PostAggregation = "reactions_count"
	// RepliesCount is the number of replies a post has received
	RepliesCount PostAggregation = "replies_count"
	// IsReview is a boolean aggregation, that is used to determine
	// whether the post is a review
	IsReview PostAggregation = "is_review"
	// AuthorSeparationDegree is the number of degrees of separation
	// between the author of the post and the current user
	AuthorSeparationDegree PostAggregation = "author_separation_degree"
	// PostDate is the date when the post was made
	PostDate PostAggregation = "post_date"
)

// nolint: gochecknoglobals
// If we could define a ListAll generic method
// we wouldn't need global lists like this
// But a method receiver can't have type parameters
var PostAggregations = []PostAggregation{
	ReactionsCount, RepliesCount, IsReview, AuthorSeparationDegree, PostDate,
}

const (
	// WorksCount is the number of works an artist has made
	// or if it's a label/studio/publishing house etc, released
	// It is non-discriminatory, so for example it can list
	// films where an actor hasn't been included in the credits
	WorksCount ArtistAggregation = "works_count"
	// MainWorksCount is the number of works an artist has made
	// or played a significant role in making
	MainWorksCount ArtistAggregation = "main_works_count"
	// ArtistType is the type of the artist.
	// Can be 'individual', 'group', 'studio'
	ArtistType ArtistAggregation = "artist_type"
	// Active is a boolean aggregation, that is used to determine
	// whether the artist is active
	Active ArtistAggregation = "active"
	// AverageWorksRating is the average rating of an artist's works
	AverageWorksRating ArtistAggregation = "average_rating"
	// AverageWorksRatingByDate is the average rating of an artist's works
	// by date range
	AverageWorksRatingByDate ArtistAggregation = "average_rating_by_date"
	// TopGenres is the list of most common genres of the artist's works
	TopGenres ArtistAggregation = "top_genres"
	// AssociatedActs is the list of projects the artist has been associated with
	AssociatedActs ArtistAggregation = "associated_acts"
	// Location is the location of the artist
	// It applies to both the birthplace and the current residence
	Location ArtistAggregation = "location"
	// ReviewsReceived is the number of reviews the artist has received
	ReviewsReceived ArtistAggregation = "reviews_received"
	// RewardsReceived is the number of rewards the artist has received
	RewardsReceived ArtistAggregation = "rewards_received"
)

// nolint: gochecknoglobals
var ArtistAggregations = []ArtistAggregation{
	WorksCount, MainWorksCount, ArtistType, Active, AverageWorksRating,
	AverageWorksRatingByDate, TopGenres, AssociatedActs, Location,
	RewardsReceived, ReviewsReceived,
}

type AnyAggregation interface {
	ArtistAggregation | PostAggregation | UserAggregation | MediaAggregation
	String() string
}

type Aggregation interface {
	String() string
}

func FromStringSlice(agg []string) (res []interface{}) {
	userAggregations := lo.Map(UserAggregations, func(a UserAggregation, _ int) string {
		return a.String()
	})
	artistAggregations := lo.Map(ArtistAggregations, func(a ArtistAggregation, _ int) string {
		return a.String()
	})
	mediaAggregations := lo.Map(MediaAggregations, func(a MediaAggregation, _ int) string {
		return a.String()
	})
	var (
		MediaAggregations []MediaAggregation
		UserAggregations  []UserAggregation
		PostAggregations  []PostAggregation
		AristAggregations []ArtistAggregation
	)
	for _, a := range agg {
		switch {
		case lo.Contains(userAggregations, a):
			UserAggregations = append(UserAggregations, UserAggregation(a))
		case lo.Contains(artistAggregations, a):
			AristAggregations = append(AristAggregations, ArtistAggregation(a))
		case lo.Contains(mediaAggregations, a):
			MediaAggregations = append(MediaAggregations, MediaAggregation(a))
		default:
			PostAggregations = append(PostAggregations, PostAggregation(a))
		}
	}
	res = []any{
		MediaAggregations, UserAggregations,
		PostAggregations, AristAggregations,
	}

	return res
}

func (s MediaAggregation) String() string {
	return string(s)
}

func (s UserAggregation) String() string {
	return string(s)
}

func (s PostAggregation) String() string {
	return string(s)
}

func (s ArtistAggregation) String() string {
	return string(s)
}
