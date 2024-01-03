package media

import (
	"context"
	"database/sql"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/rs/zerolog"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/models"
)

type (
	// IController is the interface for the media controller
	// It defines the methods that the media controller must implement
	// This is useful for mocking the media controller in unit tests
	IController interface {
		GetMedia(c *fiber.Ctx) error
		GetRandom(c *fiber.Ctx) error
		AddMedia(c *fiber.Ctx) error
	}

	// Controller is the controller for media endpoints
	// The methods which are the receivers of this struct are a bridge between the fiber layer and the storage layer
	Controller struct {
		storage models.MediaStorage
	}

	mediaError struct {
		ID  uuid.UUID
		Err error
	}
)

func NewController(storage models.MediaStorage) *Controller {
	return &Controller{storage: storage}
}

// GetMedia retrieves media information based on the media ID
// media ID is a UUID (binary, but passed from the fronetend as a string,
// since typescript doesn't support binary)
func (mc *Controller) GetMedia(c *fiber.Ctx) error {
	mediaID, err := uuid.FromString(c.Params("id"))
	if err != nil {
		mc.storage.Log.Error().Err(err).
			Msgf("Failed to parse media ID %s", c.Params("id"))
		return h.Res(c, fiber.StatusBadRequest, "Invalid media ID")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	media, err := mc.storage.Get(ctx, mediaID)
	if err != nil {
		mc.storage.Log.Error().Err(err).
			Msgf("Failed to get media with ID %s", c.Params("id"))
		return h.Res(c, fiber.StatusInternalServerError, "Failed to get media")
	}

	detailedMedia, err := mc.storage.
		GetMediaDetails(ctx, media.Kind, media.ID)
	if err != nil {
		mc.storage.Log.Error().Err(err).Msgf("Failed to get media details for media with ID %s: %v", c.Params("id"), err)
		return h.Res(c, fiber.StatusInternalServerError, "Failed to get media details")
	}

	return h.ResData(c, fiber.StatusOK, "success", detailedMedia)
}

// TODO: when upload form is implemented, flush the redis cache, since the response might change
func (mc *Controller) GetImagePaths(c *fiber.Ctx) error {
	mc.storage.Log.Info().Msg("Hit endpoint " + c.Path())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mediaID, err := uuid.FromString(c.Params("media_id"))
	if err != nil {
		return handleBadRequest(mc.storage.Log, c, "Invalid media ID")
	}

	kind, err := mc.storage.GetKind(ctx, mediaID)
	if err != nil {
		return handleInternalError(mc.storage.Log, c, "Failed to get kind")
	}
	mc.storage.Log.Debug().Msgf("Got kind %s for media with ID %s", kind, c.Params("media_id"))

	path, err := mc.storage.GetImagePath(ctx, mediaID)
	if err == sql.ErrNoRows {
		return handlePlaceholderImage(mc.storage.Log, c, kind)
	} else if err != nil {
		return handleInternalError(mc.storage.Log, c, "Failed to get image paths")
	}

	return handleImageResponse(mc.storage.Log, c, path)
}

func handleBadRequest(log *zerolog.Logger, c *fiber.Ctx, message string) error {
	log.Error().Msgf("Failed to parse media ID %s", c.Params("media_id"))
	return h.Res(c, fiber.StatusBadRequest, message)
}

func handleInternalError(log *zerolog.Logger, c *fiber.Ctx, message string) error {
	log.Error().Msg(message)
	return h.Res(c, fiber.StatusInternalServerError, message)
}

func handlePlaceholderImage(log *zerolog.Logger, c *fiber.Ctx, kind string) error {
	log.Warn().Msgf("Using placeholder image for media with ID %s", c.Params("id"))
	var placeholderPath string
	switch kind {
	case "film", "tv_show":
		placeholderPath = "./static/video/placeholder.svg"
	case "album", "track":
		placeholderPath = "./static/music/placeholder.webp"
	default:
		placeholderPath = "./static/placeholder.png"
	}

	err := c.SendString(placeholderPath)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to send placeholder image for media with ID %s", c.Params("id"))
		return h.Res(c, fiber.StatusNotFound, "Failed to send placeholder image")
	}
	return c.SendStatus(fiber.StatusOK)
}

func handleImageResponse(log *zerolog.Logger, c *fiber.Ctx, path string) error {
	log.Debug().Msgf("Got image path %s for media with ID %s", path, c.Params("id"))
	err := c.SendString("./static/" + path)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to send image for media with ID %s", c.Params("id"))
		return h.Res(c, fiber.StatusNotFound, "Failed to send image")
	}
	return c.SendStatus(fiber.StatusOK)
}
