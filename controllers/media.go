package controllers

import (
	"context"
	"database/sql"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"

	"codeberg.org/mjh/LibRate/internal/client"
	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/models"
	services "codeberg.org/mjh/LibRate/recommendation/go/services"
)

type (
	// IMediaController is the interface for the media controller
	// It defines the methods that the media controller must implement
	// This is useful for mocking the media controller in unit tests
	IMediaController interface {
		GetMedia(c *fiber.Ctx) error
		GetRandom(c *fiber.Ctx) error
		AddMedia(c *fiber.Ctx) error
	}

	// MediaController is the controller for media endpoints
	// The methods which are the receivers of this struct are a bridge between the fiber layer and the storage layer
	MediaController struct {
		storage models.MediaStorage
	}

	mediaError struct {
		ID  uuid.UUID
		Err error
	}
)

func NewMediaController(storage models.MediaStorage) *MediaController {
	return &MediaController{storage: storage}
}

// GetMedia retrieves media information based on the media ID
// media ID is a UUID (binary, but passed from the fronetend as a string,
// since typescript doesn't support binary)
func (mc *MediaController) GetMedia(c *fiber.Ctx) error {
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
		mc.storage.Log.Error().Err(err).Msgf("Failed to get media details for media with ID %s: %w", c.Params("id"), err)
		return h.Res(c, fiber.StatusInternalServerError, "Failed to get media details")
	}

	return h.ResData(c, fiber.StatusOK, "success", detailedMedia)
}

// GetRandom fetches up to 5 random media items to be displayed in a carousel on the home page
func (mc *MediaController) GetRandom(c *fiber.Ctx) error {
	mc.storage.Log.Info().Msg("Hit endpoint " + c.Path())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// FIXME: blacklist tracks, because they get e.g. improperly parsed as albums at some point
	media, err := mc.storage.GetRandom(ctx, 2, "track")
	if err != nil {
		mc.storage.Log.Error().Err(err).Msgf("Failed to get random media: %s", err.Error())
		return h.Res(c, fiber.StatusInternalServerError,
			"Failed to get random media: "+err.Error())
	}

	mc.storage.Log.Info().Msgf("Got %d random media items", len(media))
	mediaItems := make([]interface{}, len(media))

	errChan := make(chan mediaError, len(media))
	var wg sync.WaitGroup

	// set an iterator variable so we can access the media item in the goroutine
	i := 0
	mc.storage.Log.Info().Msg("Getting media details")
	for id, kind := range media {
		wg.Add(1)
		go func(i int, id uuid.UUID, kind string) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			// NOTE:this val. of err must be shadowed, otherwise it will be the same err for all goroutines
			mDetails, err := mc.storage.
				GetMediaDetails(ctx, kind, id)
			if err != nil {
				errChan <- mediaError{ID: id, Err: err}
				return
			}
			mc.storage.Log.Info().Msgf("Got media details for media with ID %s", id.String())
			mediaItems[i] = mDetails
		}(i, id, kind)
		i++
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		for e := range errChan {
			mc.storage.Log.Error().Err(err).
				Msgf("Failed to get media details for media with ID %s: %s", e.ID, e.Err.Error())
		}
		return h.Res(c, fiber.StatusInternalServerError, "Failed to get media details")
	}

	return h.ResData(c, fiber.StatusOK, "success", mediaItems)
}

func (mc *MediaController) GetImagePaths(c *fiber.Ctx) error {
	mc.storage.Log.Info().Msg("Hit endpoint " + c.Path())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mediaID, err := uuid.FromString(c.Params("id"))
	if err != nil {
		mc.storage.Log.Error().Err(err).Msgf("Failed to parse media ID %s", c.Params("id"))
		return h.Res(c, fiber.StatusBadRequest, "Invalid media ID")
	}
	kind, err := mc.storage.GetKind(ctx, mediaID)
	if err != nil {
		mc.storage.Log.Error().Err(err).Msgf("Failed to get kind for media with ID %s", c.Params("id"))
		return h.Res(c, fiber.StatusInternalServerError, "Failed to get kind")
	}

	path, err := mc.storage.GetImagePath(ctx, mediaID)
	if err == sql.ErrNoRows {
		mc.storage.Log.Warn().Msgf("Using placeholder image for media with ID %s", c.Params("id"))
		switch kind {
		case "film", "tv_show":
			err = c.SendString("./static/film/placeholder.png")
			if err != nil {
				mc.storage.Log.Error().Err(err).Msgf("Failed to send placeholder image for media with ID %s", c.Params("id"))
				return h.Res(c, fiber.StatusNotFound, "Failed to send placeholder image")
			}
			return c.SendStatus(fiber.StatusOK)
		case "album", "track":
			err = c.SendString("./static/music/placeholder.webp")
			if err != nil {
				mc.storage.Log.Error().Err(err).Msgf("Failed to send placeholder image for media with ID %s", c.Params("id"))
				return h.Res(c, fiber.StatusNotFound, "Failed to send placeholder image")
			}
			return c.SendStatus(fiber.StatusOK)
		default:
			err = c.SendString("./static/placeholder.png")
			if err != nil {
				mc.storage.Log.Error().Err(err).Msgf("Failed to send placeholder image for media with ID %s", c.Params("id"))
				return h.Res(c, fiber.StatusNotFound, "Failed to send placeholder image")
			}
			return c.SendStatus(fiber.StatusOK)
		}
	} else if err != nil {
		mc.storage.Log.Error().Err(err).Msgf("Failed to get image paths for media with ID %s", c.Params("id"))
		return h.Res(c, fiber.StatusInternalServerError, "Failed to get image paths")
	}
	mc.storage.Log.Debug().Msgf("Got image path %s for media with ID %s", path, c.Params("id"))
	err = c.SendString("./static/" + path)
	if err != nil {
		mc.storage.Log.Error().Err(err).Msgf("Failed to send image for media with ID %s", c.Params("id"))
		return h.Res(c, fiber.StatusNotFound, "Failed to send image")
	}
	return c.SendStatus(fiber.StatusOK)
}

// WARN: this is probably wrong
// GetRecommendations returns media recommendations for a user based on collaborative filtering
// TODO: the actual underlying functionality, i.e. the recommendations server
func GetRecommendations(c *fiber.Ctx) error {
	mID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		//nolint:errcheck
		return h.Res(c, fiber.StatusBadRequest, fmt.Sprintf("Invalid member ID %s (must be an integer)", c.Params("id")))
	}

	memberID := int32(mID)

	conn, err := client.ConnectToService(context.Background(), "recommendation", "50051")
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, "Failed to connect to recommendation service")
	}
	defer conn.Close()

	s := services.NewRecommendationServiceClient(conn)

	recommendedMedia, err := s.GetRecommendations(context.Background(), &services.GetRecommendationsRequest{
		MemberId: memberID,
	})
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, "Failed to get recommendations")
	}

	return c.JSON(recommendedMedia)
}

func AddGenre() {
	// TODO: implement
}
