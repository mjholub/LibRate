package static

import (
	"codeberg.org/mjh/LibRate/cfg"
	models "codeberg.org/mjh/LibRate/models/static"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type StaticController struct {
	storage *models.Storage
	log     *zerolog.Logger
	conf    *cfg.Config
}

func NewStaticController(conf *cfg.Config, dbConn *sqlx.DB, logger *zerolog.Logger) *StaticController {
	return &StaticController{
		conf:    conf,
		log:     logger,
		storage: models.NewStorage(dbConn, logger),
	}
}
