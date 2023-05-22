package models

import (
	"context"
	"fmt"
	"time"

	"codeberg.org/mjh/LibRate/cfg"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

type RatingInput struct {
	NumStars    int    `json:"numstars"`
	Comment     string `json:"comment", omitempty`
	Topic       string `json:"topic", omitempty`
	Attribution string `json:"attribution", omitempty`
	UserID      int    `json:"userid"`
	MediaID     int    `json:"mediaid"`
}

type Rating struct {
	NumStars    int    `json:"numstars"`
	Comment     string `json:"comment", omitempty`
	Topic       string `json:"topic", omitempty`
	Attribution string `json:"attribution", omitempty`
	UserID      int    `json:"userid"`
	MediaID     int    `json:"mediaid"`
}

type RatingStorer interface {
	SaveRating(rating *Rating) error
	Get(ctx context.Context, key string) (*Rating, error)
	GetAll() ([]*Rating, error)
}

type RatingStorage struct{}

func NewRatingStorage() *RatingStorage {
	return &RatingStorage{}
}

func (rs *RatingStorage) SaveRating(rating *Rating) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	config := cfg.LoadConfig()
	dbConf := config.ArangoDB

	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{fmt.Sprintf("http://%s:%s", dbConf.Host, dbConf.Port)},
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

	config := cfg.LoadConfig()
	dbConf := config.ArangoDB

	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{fmt.Sprintf("http://%s:%s", dbConf.Host, dbConf.Port)},
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
