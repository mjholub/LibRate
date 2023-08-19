package form

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/models"
)

type (
	FormController struct {
		log     *zerolog.Logger
		storage models.MediaStorage
	}
)

func NewFormController(log *zerolog.Logger, storage models.MediaStorage) *FormController {
	return &FormController{
		log:     log,
		storage: storage,
	}
}

func (fc *FormController) AddMedia(c *fiber.Ctx) error {
	mediaType := c.Params("type")
	switch mediaType {
	case "film":
		err := fc.addFilm(c)
		if err != nil {
			return err
		}
	default:
		return h.Res(c, fiber.StatusBadRequest,
			"Sorry, adding this media type via Web UI is not supported yet")
	}

	return h.Res(c, fiber.StatusOK,
		`Media added successfully. Thank you for your contribution and please wait for an approval!
		<a href="/form/add_media">Add another media</a>`)
}

func (fc *FormController) UpdateMedia(c *fiber.Ctx) error {
	mediaType := c.Params("type")
	switch mediaType {
	case "film":
		err := fc.updateFilm(c)
		if err != nil {
			return err
		}
	default:
		return h.Res(c, fiber.StatusBadRequest,
			"Sorry, updating this media type via Web UI is not supported yet")
	}

	return h.Res(c, fiber.StatusOK,
		`Media updated successfully. Thank you for your contribution and please wait for an approval!
		<a href="/form/update_media">Update another media</a>`)
}

func (fc *FormController) addFilm(c *fiber.Ctx) (err error) {
	var film *models.Film
	if err := c.BodyParser(&film); err != nil {
		fc.log.Error().Err(err).Msgf("Failed to parse JSON: %s", err.Error())
		return h.Res(c, fiber.StatusBadRequest, "Cannot parse JSON")
	}

	fc.storage.AddFilm(c.UserContext(), film)

	return nil
}

func (fc *FormController) updateFilm(c *fiber.Ctx) (err error) {
	var film *models.Film
	if err := c.BodyParser(&film); err != nil {
		fc.log.Error().Err(err).Msgf("Failed to parse JSON: %s", err.Error())
		return h.Res(c, fiber.StatusBadRequest, "Cannot parse JSON")
	}

	fc.storage.UpdateFilm(c.UserContext(), film)

	return nil
}
