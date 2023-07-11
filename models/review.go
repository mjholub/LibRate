package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type (
	RatingInput struct {
		// TODO: allow for setting dynamic rating scales
		NumStars    uint8     `json:"numstars" binding:"required" validate:"min=1,max=10" error:"numstars must be between 1 and 10" db:"stars"`
		Comment     string    `json:"comment,omitempty" db:"comment"`
		Topic       string    `json:"topic,omitempty" db:"topic"`
		Attribution string    `json:"attribution,omitempty" db:"attribution"`
		UserID      uint32    `json:"userid" db:"user_id"`
		MediaID     uuid.UUID `json:"mediaid" db:"media_id"`
	}

	Rating struct {
		ID          int64     `json:"_key" db:"id,pk"`
		CreatedAt   time.Time `json:"created_at" db:"created_at"`
		NumStars    uint8     `json:"numstars" binding:"required" validate:"min=1,max=10" error:"numstars must be between 1 and 10" db:"stars" `
		Comment     string    `json:"comment,omitempty" db:"comment"`
		Topic       string    `json:"topic,omitempty" db:"topic"`
		Attribution string    `json:"attribution,omitempty" db:"attribution"`
		UserID      uint32    `json:"userid" db:"user_id"`
		MediaID     uuid.UUID `json:"mediaid" db:"media_id"`
		// track/cast/theme
		TrackRatings *TrackRating `json:"trackRatings,omitempty" db:"track_rating"`
		CastRating   *CastRating  `json:"castRating,omitempty" db:"cast_rating"`
	}

	UpdateableKeyTypes interface {
		~int | ~uint | string
	}

	/*
	* It should probably be better from the perspective of the UX
	* as well as the performance, normalization, modularity and reusability
	* to have a separate table for each kind of rating
	 */

	TrackRating struct {
		ID       int64  `json:"_key" db:"id,pk"`
		Track    *Track `json:"track" db:"track"`
		NumStars uint8  `json:"numstars" binding:"required" validate:"min=1,max=10" error:"numstars must be between 1 and 10" db:"stars" `
		UserID   uint32 `json:"userid" db:"user_id"`
	}

	CastRating struct {
		ID       int64  `json:"_key" db:"id,pk"`
		Cast     *Cast  `json:"cast" db:"cast_id"`
		NumStars uint8  `json:"numstars" binding:"required" validate:"min=1,max=10" error:"numstars must be between 1 and 10" db:"stars" `
		UserID   uint32 `json:"userid" db:"user_id"`
	}

	// Update is not present, because methods cannot have type parameters
	RatingStorer interface {
		New(ri *RatingInput) error
		Get(ctx context.Context, ID int64) (*Rating, error)
		GetAll() ([]*Rating, error)
		GetByMediaID(ctx context.Context, mediaID uuid.UUID) ([]*Rating, error)
	}

	RatingStorage struct {
		db  *sqlx.DB
		log *zerolog.Logger
	}
)

func NewRatingStorage(db *sqlx.DB, log *zerolog.Logger) *RatingStorage {
	return &RatingStorage{}
}

func (rs *RatingStorage) New(ctx context.Context, rating *RatingInput) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		stmt, err := rs.db.PreparexContext(ctx,
			`INSERT INTO reviews.ratings (stars, comment, topic, attribution, user_id, media_id)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`)
		if err != nil {
			return fmt.Errorf("error preparing statement: %w", err)
		}
		defer stmt.Close()

		var id int64

		err = stmt.QueryRowxContext(ctx,
			rating.NumStars,
			rating.Comment,
			rating.Topic,
			rating.Attribution,
			rating.UserID,
			rating.MediaID,
		).Scan(&id)

		if err != nil {
			return fmt.Errorf("error inserting rating: %w", err)
		}
		rs.log.Debug().Msgf("Inserted rating with id %d", id)

		return nil
	}
}

func UpdateRating[U UpdateableKeyTypes](ctx context.Context, rs *RatingStorage, id int64, values []U) (err error) {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:

		for v := range values {
			_, err = rs.db.ExecContext(ctx, `UPDATE reviews.ratings SET $1 = $2 WHERE id = $3`, v, values[v], id)
			if err != nil {
				return fmt.Errorf("error updating rating: %w", err)
			}
		}
		return nil
	}
}

// Get retrieves a rating by its id.
func (rs *RatingStorage) Get(ctx context.Context, id int64) (r Rating, err error) {
	err = rs.db.GetContext(ctx, &r, `SELECT * FROM reviews.ratings WHERE id = $1`, id)
	if err != nil {
		return Rating{}, fmt.Errorf("error getting rating: %w", err)
	}
	return r, nil
}

// GetLatestRatings retrieves the latest reviews for all media items. The limit and offset
// parameters are used for pagination.
func (rs *RatingStorage) GetLatest(ctx context.Context, limit int, offset int) (ratings []*Rating, err error) {
	err = rs.db.SelectContext(ctx, &ratings, `SELECT * FROM reviews.ratings 
		ORDER BY created_at
		DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error getting ratings: %w", err)
	}
	return ratings, nil
}

func (rs *RatingStorage) GetAll() (ratings []*Rating, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = rs.db.SelectContext(ctx, &ratings, `SELECT * FROM reviews.ratings`)
	if err != nil {
		return nil, fmt.Errorf("error getting ratings: %w", err)
	}
	return ratings, nil
}

func (rs *RatingStorage) GetByMediaID(ctx context.Context, mediaID uuid.UUID) (ratings []*Rating, err error) {
	err = rs.db.SelectContext(
		ctx, &ratings, `SELECT * FROM reviews.ratings WHERE media_id = $1`, mediaID)
	if err != nil {
		return nil, fmt.Errorf("error getting ratings: %w", err)
	}
	return ratings, nil
}

func (rs *RatingStorage) GetAverageStars(ctx context.Context, rating interface{},
	mediaID uuid.UUID,
) (avgStars float64, err error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		var avgStarsFloat sql.NullFloat64

		switch rating.(type) {
		case *Track:
			err = rs.db.GetContext(ctx, &avgStarsFloat,
				`SELECT AVG(stars) FROM reviews.track_ratings WHERE track_id = $1`, mediaID)
			if err != nil {
				return 0, fmt.Errorf("error getting average stars: %w", err)
			}
		case *CastRating:
			err = rs.db.GetContext(ctx, &avgStarsFloat,
				`SELECT AVG(stars) FROM reviews.cast_ratings WHERE cast_id = $1`, mediaID)
			if err != nil {
				return 0, fmt.Errorf("error getting average stars: %w", err)
			}
		case *Rating:
			err = rs.db.GetContext(ctx, &avgStarsFloat,
				`SELECT AVG(stars) FROM reviews.ratings WHERE media_id = $1`, mediaID)
			if err != nil {
				return 0, fmt.Errorf("error getting average stars: %w", err)
			}
		default:
			return 0, fmt.Errorf("invalid type")
		}

		return avgStarsFloat.Float64, nil
	}
}
