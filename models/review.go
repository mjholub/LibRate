package models

import (
	"context"
	"fmt"
	"time"

	"codeberg.org/mjh/LibRate/cfg"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/gofrs/uuid/v5"
)

type (
	RatingInput struct {
		// TODO: allow for setting dynamic rating scales
		NumStars    uint8  `json:"numstars" binding:"required" validate:"min=1,max=10" error:"numstars must be between 1 and 10" db:"stars"`
		Comment     string `json:"comment,omitempty" db:"comment"`
		Topic       string `json:"topic,omitempty" db:"topic"`
		Attribution string `json:"attribution,omitempty" db:"attribution"`
		UserID      uint32 `json:"userid" db:"user_id"`
		MediaID     uint   `json:"mediaid" db:"media_id"`
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
		ID       int64     `json:"_key" db:"id,pk"`
		MediaID  uuid.UUID `json:"mediaid" db:"media_id"`
		Cast     *Cast     `json:"cast" db:"cast"`
		NumStars uint8     `json:"numstars" binding:"required" validate:"min=1,max=10" error:"numstars must be between 1 and 10" db:"stars" `
		UserID   uint32    `json:"userid" db:"user_id"`
	}

	// Theme vote serves as the basis for constructing most relevant tags for a given media
	ThemeVote struct {
		ID       int64     `json:"_key" db:"id,pk"`
		MediaID  uuid.UUID `json:"mediaid" db:"media_id"`
		Theme    string    `json:"theme" db:"theme"`
		NumStars uint8     `json:"numstars" binding:"required" validate:"min=1,max=10" error:"numstars must be between 1 and 10" db:"stars" `
		UserID   uint32    `json:"userid" db:"user_id"`
	}

	RatingStorer interface {
		SaveRating(rating *Rating) error
		Get(ctx context.Context, key string) (*Rating, error)
		GetAll() ([]*Rating, error)
	}

	RatingStorage struct{}
)

func NewRatingStorage() *RatingStorage {
	return &RatingStorage{}
}

func (rs *RatingStorage) SaveRating(rating *Rating) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	config := cfg.LoadConfig().OrElse(cfg.ReadDefaults())
	dbConf := config.DBConfig

	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{fmt.Sprintf("http://%s:%d", dbConf.Host, dbConf.Port)},
	})
	if err != nil {
		return err
	}

	client, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(dbConf.User, dbConf.Password),
	})
	if err != nil {
		return err
	}

	db, err := client.Database(ctx, dbConf.Database)
	if err != nil {
		return err
	}

	ratings, err := db.Collection(ctx, "ratings")
	if err != nil {
		return err
	}

	meta, err := ratings.CreateDocument(ctx, rating)
	if err != nil {
		return err
	}

	fmt.Printf("Created document with key: %s\n", meta.Key)
	return nil
}

func (rs *RatingStorage) Get(ctx context.Context, key interface{}) (*Rating, error) {
	var rating Rating

	config := cfg.LoadConfig().OrElse(cfg.ReadDefaults())
	dbConf := config.DBConfig

	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{fmt.Sprintf("http://%s:%d", dbConf.Host, dbConf.Port)},
	})
	if err != nil {
		return nil, err
	}

	client, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(dbConf.User, dbConf.Password),
	})
	if err != nil {
		return nil, err
	}

	db, err := client.Database(ctx, dbConf.Database)
	if err != nil {
		return nil, err
	}

	ratings, err := db.Collection(ctx, "ratings")
	if err != nil {
		return nil, err
	}
	_ = ratings

	ratingKey := fmt.Sprintf("ratings/%s", key)

	ratingDoc, err := ratings.ReadDocument(ctx, ratingKey, &rating)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Read document with key '%s' from collection '%s': %s\n", ratingDoc.Key, "ratings", ratingDoc)

	return &rating, nil
}

func (rs *RatingStorage) GetPinned(ctx context.Context) ([]*Rating, error) {
	var ratings []*Rating

	config := cfg.LoadConfig().OrElse(cfg.ReadDefaults())
	dbConf := config.DBConfig

	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{fmt.Sprintf("http://%s:%d", dbConf.Host, dbConf.Port)},
	})
	if err != nil {
		return nil, err
	}

	client, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(dbConf.User, dbConf.Password),
	})
	if err != nil {
		return nil, err
	}

	db, err := client.Database(ctx, dbConf.Database)
	if err != nil {
		return nil, err
	}

	// FIXME: This is a hack to get around the fact that the collection is not created
	_, err = db.Collection(ctx, "ratings")
	if err != nil {
		return nil, err
	}

	query := "FOR r IN ratings FILTER r.pinned == true RETURN r"
	cursor, err := db.Query(ctx, query, nil)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	for {
		var rating Rating
		_, err := cursor.ReadDocument(ctx, &rating)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}

		ratings = append(ratings, &rating)
	}

	return ratings, nil
}
