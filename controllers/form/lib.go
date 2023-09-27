package form

import (
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/models"
)

type (
	Controller struct {
		log     *zerolog.Logger
		storage models.MediaStorage
		conf    *cfg.Config
	}
)

func NewController(log *zerolog.Logger,
	storage models.MediaStorage,
	conf *cfg.Config,
) *Controller {
	return &Controller{
		log:     log,
		storage: storage,
		conf:    conf,
	}
}
