package form

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
	h "codeberg.org/mjh/LibRate/internal/handlers"
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

func (fc *Controller) AddMedia(c *fiber.Ctx) error {
	mediaType := c.Params("type")
	switch mediaType {
	case "film":
		err := fc.addFilm(c)
		if err != nil {
			return err
		}
	case "book":

	default:
		return h.Res(c, fiber.StatusNotImplemented,
			"Sorry, adding this media type via Web UI is not supported yet")
	}

	return h.Res(c, fiber.StatusOK,
		`Media added successfully. Thank you for your contribution and please wait for an approval!
		<a href="/form/add_media">Add another media</a>`)
}

func (fc *Controller) UpdateMedia(c *fiber.Ctx) error {
	mediaType := c.Params("type")
	switch mediaType {
	case "film":
		err := fc.updateFilm(c)
		if err != nil {
			return err
		}
	default:
		return h.Res(c, fiber.StatusNotImplemented,
			"Sorry, updating this media type via Web UI is not supported yet")
	}

	return h.Res(c, fiber.StatusOK,
		`Media updated successfully. Thank you for your contribution and please wait for an approval!
		<a href="/form/update_media">Update another media</a>`)
}

func (fc *Controller) addFilm(c *fiber.Ctx) (err error) {
	var film *models.Film
	if err = c.BodyParser(&film); err != nil {
		fc.log.Error().Err(err).Msgf("Failed to parse JSON: %s", err.Error())
		return h.Res(c, fiber.StatusBadRequest, "Cannot parse JSON")
	}

	err = fc.storage.AddFilm(c.UserContext(), film)
	if err != nil {
		fc.log.Error().Err(err).Msgf("Failed to add film: %s", err.Error())
		return h.Res(c, fiber.StatusInternalServerError, "Failed to add film")
	}

	return nil
}

func (fc *Controller) addBook(c *fiber.Ctx) (err error) {
	var book *models.Book
	if err = c.BodyParser(&book); err != nil {
		fc.log.Error().Err(err).Msgf("Failed to parse JSON: %s", err.Error())
		return h.Res(c, fiber.StatusBadRequest, "Cannot parse JSON")
	}

	err = fc.storage.AddBook(c.UserContext(), book)
	if err != nil {
		fc.log.Error().Err(err).Msgf("Failed to add book: %s", err.Error())
		return h.Res(c, fiber.StatusInternalServerError, "Failed to add book")
	}

	return nil
}

func (fc *Controller) updateFilm(c *fiber.Ctx) (err error) {
	var film *models.Film
	if err = c.BodyParser(&film); err != nil {
		fc.log.Error().Err(err).Msgf("Failed to parse JSON: %s", err.Error())
		return h.Res(c, fiber.StatusBadRequest, "Cannot parse JSON")
	}

	err = fc.storage.UpdateFilm(c.UserContext(), film)
	if err != nil {
		fc.log.Error().Err(err).Msgf("Failed to update film: %s", err.Error())
		return h.Res(c, fiber.StatusInternalServerError, "Failed to update film")
	}

	return nil
}
