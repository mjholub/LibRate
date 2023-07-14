package controllers

import (
	"context"
	"fmt"
	"strconv"
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
		h.Res(c, fiber.StatusBadRequest, "Invalid media ID")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	media, err := mc.storage.Get(ctx, mediaID)
	if err != nil {
		h.Res(c, fiber.StatusInternalServerError, "Failed to get media")
	}

	return c.JSON(media)
}

// GetRecommendations returns media recommendations for a user based on collaborative filtering
// FIXME: the actual underlying functionality, i.e. the recommendations server is yet to be implemented
func GetRecommendations(c *fiber.Ctx) error {
	mID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		//nolint:errcheck
		return h.Res(c, fiber.StatusBadRequest, fmt.Sprintf("Invalid member ID %s (must be an integer)", c.Params("id")))
	}

	memberID := int32(mID)

	conn, err := client.ConnectToService(context.Background(), "recommendation", "50051")
	if err != nil {
		//nolint:errcheck
		return h.Res(c, fiber.StatusInternalServerError, "Failed to connect to recommendation service")
	}
	defer conn.Close()

	s := services.NewRecommendationServiceClient(conn)

	recommendedMedia, err := s.GetRecommendations(context.Background(), &services.GetRecommendationsRequest{
		MemberId: memberID,
	})
	if err != nil {
		//nolint:errcheck
		return h.Res(c, fiber.StatusInternalServerError, "Failed to get recommendations")
	}

	return c.JSON(recommendedMedia)
}

// GetRandom fetches up to 5 random media items to be displayed in a carousel on the home page
func (mc *MediaController) GetRandom(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	media, err := mc.storage.GetRandom(ctx, 5)
	if err != nil {
		//nolint:errcheck
		mc.storage.Log.Error().Err(err).Msgf("Failed to get random media: %s", err.Error())
		h.Res(c, fiber.StatusInternalServerError, "Failed to get random media: "+err.Error())
	}

	return c.JSON(media)
}

// WARN: this is probably wrong
func (mc *MediaController) AddMedia(c *fiber.Ctx) error {
	var (
		media models.MediaService // NOTE: this is a hack to get around the fact that we can't use an interface as a parameter to c.BodyParser
		props models.Media
	)

	mediaType := c.Params("type")
	switch mediaType {
	case "film":
		var film models.Film
		if err := c.BodyParser(&film); err != nil {
			mc.storage.Log.Error().Err(err).Msgf("Failed to parse JSON: %s", err.Error())
			return h.Res(c, fiber.StatusBadRequest, "Cannot parse JSON")
		}
		media = &film
		props = models.Media{ID: *film.MediaID, Title: film.Title}
	case "album":
		var album models.Album
		if err := c.BodyParser(&album); err != nil {
			mc.storage.Log.Error().Err(err).Msgf("Failed to parse JSON: %s", err.Error())
			return h.Res(c, fiber.StatusBadRequest, "Cannot parse JSON")
		}
		media = &album
		props = models.Media{ID: *album.MediaID, Title: album.Name}
	case "track":
		var track models.Track
		if err := c.BodyParser(&track); err != nil {
			mc.storage.Log.Error().Err(err).Msgf("Failed to parse JSON: %s", err.Error())
			return h.Res(c, fiber.StatusBadRequest, "Cannot parse JSON")
		}
		media = &track
		props = models.Media{ID: *track.MediaID, Title: track.Name}
	case "book":
		var book models.Book
		if err := c.BodyParser(&book); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}
		media = &book
		props = models.Media{ID: *book.MediaID, Title: book.Title}
	case "tvshow":
		var tvshow models.TVShow
		if err := c.BodyParser(&tvshow); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}
		media = &tvshow
		props = models.Media{ID: *tvshow.MediaID, Title: tvshow.Title}
	case "season":
		var season models.Season
		if err := c.BodyParser(&season); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}
		media = &season
		props = models.Media{ID: *season.MediaID, Title: strconv.Itoa(int(season.Number))}
	case "episode":
		var episode models.Episode
		if err := c.BodyParser(&episode); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}
		media = &episode
		props = models.Media{ID: *episode.MediaID, Title: episode.Title}
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid media type",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := mc.storage.Add(ctx, nil, media, props) // TODO: marshal into key-value pairs
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add media",
		})
	}

	return c.JSON(media)
}

func AddGenre() {
	// TODO: implement
}
