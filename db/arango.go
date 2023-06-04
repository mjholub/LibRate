package db

import (
	"context"
	"errors"
	"os"

	"codeberg.org/mjh/LibRate/internal/logging"

	driver "github.com/arangodb/go-driver"

	"github.com/arangodb/go-driver/http"
)

var (
	db                               driver.Database
	mediaCol, reviewsCol, membersCol driver.Collection
)

type Database interface {
	Init() error
	CreateDocument(collection string, doc interface{}) (driver.DocumentMeta, error)
	GetDocument(collection string, key string, doc interface{}) (driver.DocumentMeta, error)
	GetDocuments(collection string, query string, bindVars map[string]interface{}, doc interface{}) (driver.Cursor, error)
	UpdateDocument(collection string, key string, doc interface{}) (driver.DocumentMeta, error)
	DeleteDocument(collection string, key string) (driver.DocumentMeta, error)
}

type DatabaseImpl struct{}

func (d *DatabaseImpl) Init() error {
	log := logging.Init()
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{"http://localhost:8529"},
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create database connection")
		return err
	}

	switch env := os.Getenv("ENV"); env {
	case "dev":
		client, err := driver.NewClient(driver.ClientConfig{
			Connection:     conn,
			Authentication: driver.BasicAuthentication("root", "root"),
		})
		if err != nil {
			log.Panic().Err(err).Msg("Failed to create database client")
			return err
		}

		// Load the (main) media ratings collection
		db, err = client.Database(context.Background(), "MediaRatings")
		if err != nil {
			log.Panic().Err(err).Msg("Failed to get database")
			return err
		}

		// Load the media collection (media metadata)
		mediaCol, err = db.Collection(context.Background(), "Media")
		if err != nil {
			log.Error().Err(err).Msg("Failed to get media collection")
			return err
		}

		// Load the reviews collection (member reviews)
		reviewsCol, err = db.Collection(context.Background(), "Reviews")
		if err != nil {
			log.Error().Err(err).Msg("Failed to get reviews collection")
			return err
		}

		// Load the members collection (member metadata)
		membersCol, err = db.Collection(context.Background(), "members")
		if err != nil {
			log.Error().Err(err).Msg("Failed to get members collection")
			return err
		}
	default:
		log.Panic().Str("env", env).Msg("Invalid environment")
		return errors.New("Invalid environment. LibRate is not production ready yet!")
	}
	return nil
}

func (d *DatabaseImpl) CreateDocument(collection string, doc interface{}) (driver.DocumentMeta, error) {
	log := logging.Init()
	var col driver.Collection
	switch collection {
	case "Media":
		col = mediaCol
	case "Reviews":
		col = reviewsCol
	case "Members":
		col = membersCol
	default:
		log.Error().Str("collection", collection).Msg("Invalid collection")
		return driver.DocumentMeta{}, errors.New("Invalid collection")
	}
	meta, err := col.CreateDocument(context.Background(), doc)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create document")
		return driver.DocumentMeta{}, err
	}
	return meta, nil
}

func (d *DatabaseImpl) GetDocument(collection string, key string, doc interface{}) (driver.DocumentMeta, error) {
	log := logging.Init()
	var col driver.Collection
	switch collection {
	case "Media":
		col = mediaCol
	case "Reviews":
		col = reviewsCol
	case "Members":
		col = membersCol
	default:
		log.Error().Str("collection", collection).Msg("Invalid collection")
		return driver.DocumentMeta{}, errors.New("Invalid collection")
	}
	meta, err := col.ReadDocument(context.Background(), key, doc)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get document")
		return driver.DocumentMeta{}, err
	}
	return meta, nil
}

func (d *DatabaseImpl) GetDocuments(collection string, query string, bindVars map[string]interface{}, doc interface{}) (driver.Cursor, error) {
	log := logging.Init()
	var col driver.Collection
	switch collection {
	case "Media":
		col = mediaCol
	case "Reviews":
		col = reviewsCol
	case "Members":
		col = membersCol
	default:
		log.Error().Str("collection", collection).Msg("Invalid collection")
		return nil, errors.New("Invalid collection")
	}
	cursor, err := col.Database().CreateArangoSearchView(context.Background(), query, bindVars)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get documents")
		return nil, err
	}
	return cursor, nil
}

func (d *DatabaseImpl) UpdateDocument(collection string, key string, doc interface{}) (driver.DocumentMeta, error) {
	log := logging.Init()
	var col driver.Collection
	switch collection {
	case "Media":
		col = mediaCol
	case "Reviews":
		col = reviewsCol
	case "Members":
		col = membersCol
	default:
		log.Error().Str("collection", collection).Msg("Invalid collection")
		return driver.DocumentMeta{}, errors.New("Invalid collection")
	}
	meta, err := col.UpdateDocument(context.Background(), key, doc)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update document")
		return driver.DocumentMeta{}, err
	}
	return meta, nil
}

func (d *DatabaseImpl) DeleteDocument(collection string, key string) (driver.DocumentMeta, error) {
	log := logging.Init()
	var col driver.Collection
	switch collection {
	case "Media":
		col = mediaCol
	case "Reviews":
		col = reviewsCol
	case "Members":
		col = membersCol
	default:
		log.Error().Str("collection", collection).Msg("Invalid collection")
		return driver.DocumentMeta{}, errors.New("Invalid collection")
	}
	meta, err := col.RemoveDocument(context.Background(), key)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete document")
		return driver.DocumentMeta{}, err
	}
	return meta, nil
}
