package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	_ "github.com/jmoiron/sqlx"

	_ "codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/internal/client"
	"codeberg.org/mjh/LibRate/models"
	services "codeberg.org/mjh/LibRate/recommendation/go/services"
)

// GetMedia retrieves media information based on the media ID
func GetMedia(c *fiber.Ctx) error {
	rStorage := models.NewRatingStorage()
	mediaID, _ := strconv.Atoi(c.Params("id"))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	media, err := rStorage.Get(ctx, mediaID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Media not found",
		})
	}

	return c.JSON(media)
}

// GetRecommendations returns media recommendations for a user based on collaborative filtering
func GetRecommendations(c *fiber.Ctx) error {
	mID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid member ID",
		})
	}
	memberID := int32(mID)

	conn, err := client.ConnectToService(context.Background(), "recommendation", "50051")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to connect to recommendation service",
		})
	}
	defer conn.Close()

	s := services.NewRecommendationServiceClient(conn)

	recommendedMedia, err := s.GetRecommendations(context.Background(), &services.GetRecommendationsRequest{
		MemberId: memberID,
	})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get recommendations",
		})
	}

	return c.JSON(recommendedMedia)
}

func AddMedia(c *fiber.Ctx) error {
	mstor := models.NewMediaStorage()
	var (
		media models.MediaService // NOTE: this is a hack to get around the fact that we can't use an interface as a parameter to c.BodyParser
		props models.Media
	)

	mediaType := c.Params("type")
	switch mediaType {
	case "film":
		var film models.Film
		if err := c.BodyParser(&film); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}
		media = &film
		props = models.Media{UUID: *film.MediaID, Name: film.Title}
	case "album":
		var album models.Album
		if err := c.BodyParser(&album); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}
		media = &album
		props = models.Media{UUID: *album.MediaID, Name: album.Name}
	case "genre":
		var genre models.Genre
		if err := c.BodyParser(&genre); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}
		media = &genre
		props = models.Media{UUID: *genre.MediaID, Name: genre.Name}
	case "track":
		var track models.Track
		if err := c.BodyParser(&track); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}
		media = &track
		props = models.Media{UUID: *track.MediaID, Name: track.Name}
	case "book":
		var book models.Book
		if err := c.BodyParser(&book); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}
		media = &book
		props = models.Media{UUID: *book.MediaID, Name: book.Title}
	case "tvshow":
		var tvshow models.TVShow
		if err := c.BodyParser(&tvshow); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}
		media = &tvshow
		props = models.Media{UUID: *tvshow.MediaID, Name: tvshow.Title}
	case "season":
		var season models.Season
		if err := c.BodyParser(&season); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}
		media = &season
		props = models.Media{UUID: *season.MediaID, Name: strconv.Itoa(int(season.Number))}
	case "episode":
		var episode models.Episode
		if err := c.BodyParser(&episode); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}
		media = &episode
		props = models.Media{UUID: *episode.MediaID, Name: episode.Title}
	default:
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid media type",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := mstor.Add(ctx, nil, media, props) // TODO: marshal into key-value pairs
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add media",
		})
	}

	return c.JSON(media)
}
