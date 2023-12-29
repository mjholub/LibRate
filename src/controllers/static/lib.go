package static

import (
	"codeberg.org/mjh/LibRate/cfg"
	models "codeberg.org/mjh/LibRate/models/static"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type Controller struct {
	storage *models.Storage
	log     *zerolog.Logger
	conf    *cfg.Config
}

func NewController(conf *cfg.Config, dbConn *sqlx.DB, logger *zerolog.Logger) *Controller {
	return &Controller{
		conf:    conf,
		log:     logger,
		storage: models.NewStorage(dbConn, logger),
	}
}
