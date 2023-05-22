package db

import (
	"context"
	"errors"
	"os"

	"codeberg.org/mjh/LibRate/utils"

	driver "github.com/arangodb/go-driver"

	"github.com/arangodb/go-driver/http"
	"go.uber.org/zap"
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
	log := utils.NewLogger()
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{"http://localhost:8529"},
	})
	if err != nil {
		log.Panic("Failed to connect to database", zap.Error(err))
		return err
	}

	switch env := os.Getenv("ENV"); env {
	case "dev":
		client, err := driver.NewClient(driver.ClientConfig{
			Connection:     conn,
			Authentication: driver.BasicAuthentication("root", "root"),
		})
		if err != nil {
			log.Panic("Failed to create client", zap.Error(err))
			return err
		}

		// Load the (main) media ratings collection
		db, err = client.Database(context.Background(), "MediaRatings")
		if err != nil {
			log.Panic("Failed to get database", zap.Error(err))
			return err
		}

		// Load the media collection (media metadata)
		mediaCol, err = db.Collection(context.Background(), "Media")
		if err != nil {
			log.Panic("Failed to get media collection", zap.Error(err))
			return err
		}

		// Load the reviews collection (member reviews)
		reviewsCol, err = db.Collection(context.Background(), "Reviews")
		if err != nil {
			log.Panic("Failed to get reviews collection", zap.Error(err))
			return err
		}

		// Load the members collection (member metadata)
		membersCol, err = db.Collection(context.Background(), "members")
		if err != nil {
			log.Panic("Failed to get members collection", zap.Error(err))
			return err
		}
	default:
		log.Panic("Invalid environment", zap.String("env", env))
		return errors.New("Invalid environment. Librerym is not production ready yet!")
	}
	return nil
}

func (d *DatabaseImpl) CreateDocument(collection string, doc interface{}) (driver.DocumentMeta, error) {
	log := utils.NewLogger()
	var col driver.Collection
	switch collection {
	case "Media":
		col = mediaCol
	case "Reviews":
		col = reviewsCol
	case "Members":
		col = membersCol
	default:
		log.Panic("Invalid collection", zap.String("collection", collection))
		return driver.DocumentMeta{}, errors.New("Invalid collection")
	}
	meta, err := col.CreateDocument(context.Background(), doc)
	if err != nil {
		log.Panic("Failed to create document", zap.Error(err))
		return driver.DocumentMeta{}, err
	}
	return meta, nil
}

func (d *DatabaseImpl) GetDocument(collection string, key string, doc interface{}) (driver.DocumentMeta, error) {
	log := utils.NewLogger()
	var col driver.Collection
	switch collection {
	case "Media":
		col = mediaCol
	case "Reviews":
		col = reviewsCol
	case "Members":
		col = membersCol
	default:
		log.Panic("Invalid collection", zap.String("collection", collection))
		return driver.DocumentMeta{}, errors.New("Invalid collection")
	}
	meta, err := col.ReadDocument(context.Background(), key, doc)
	if err != nil {
		log.Panic("Failed to read document", zap.Error(err))
		return driver.DocumentMeta{}, err
	}
	return meta, nil
}

func (d *DatabaseImpl) GetDocuments(collection string, query string, bindVars map[string]interface{}, doc interface{}) (driver.Cursor, error) {
	log := utils.NewLogger()
	var col driver.Collection
	switch collection {
	case "Media":
		col = mediaCol
	case "Reviews":
		col = reviewsCol
	case "Members":
		col = membersCol
	default:
		log.Panic("Invalid collection", zap.String("collection", collection))
		return nil, errors.New("Invalid collection")
	}
	cursor, err := col.Database().CreateArangoSearchView(context.Background(), query, bindVars)
	if err != nil {
		log.Panic("Failed to query documents", zap.Error(err))
		return nil, err
	}
	return cursor, nil
}

func (d *DatabaseImpl) UpdateDocument(collection string, key string, doc interface{}) (driver.DocumentMeta, error) {
	log := utils.NewLogger()
	var col driver.Collection
	switch collection {
	case "Media":
		col = mediaCol
	case "Reviews":
		col = reviewsCol
	case "Members":
		col = membersCol
	default:
		log.Panic("Invalid collection", zap.String("collection", collection))
		return driver.DocumentMeta{}, errors.New("Invalid collection")
	}
	meta, err := col.UpdateDocument(context.Background(), key, doc)
	if err != nil {
		log.Panic("Failed to update document", zap.Error(err))
		return driver.DocumentMeta{}, err
	}
	return meta, nil
}

func (d *DatabaseImpl) DeleteDocument(collection string, key string) (driver.DocumentMeta, error) {
	log := utils.NewLogger()
	var col driver.Collection
	switch collection {
	case "Media":
		col = mediaCol
	case "Reviews":
		col = reviewsCol
	case "Members":
		col = membersCol
	default:
		log.Panic("Invalid collection", zap.String("collection", collection))
		return driver.DocumentMeta{}, errors.New("Invalid collection")
	}
	meta, err := col.RemoveDocument(context.Background(), key)
	if err != nil {
		log.Panic("Failed to remove document", zap.Error(err))
		return driver.DocumentMeta{}, err
	}
	return meta, nil
}
