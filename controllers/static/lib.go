package static

import (
	"github.com/gofiber/storage/redis/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
	models "codeberg.org/mjh/LibRate/models/static"
)

type Controller struct {
	storage *models.Storage
	cache   *redis.Storage
	log     *zerolog.Logger
	conf    *cfg.Config
}

func NewController(conf *cfg.Config, dbConn *pgxpool.Pool, logger *zerolog.Logger) *Controller {
	return &Controller{
		conf:    conf,
		log:     logger,
		storage: models.NewStorage(dbConn, logger),
	}
}
