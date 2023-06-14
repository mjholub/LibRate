package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

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
	var media models.MediaService // NOTE: this is a hack to get around the fact that we can't use an interface as a parameter to c.BodyParser

	mediaType := c.Params("type")
	switch mediaType {
	case "film":
		var film models.Film
		if err := c.BodyParser(&film); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}
		media = film
	case "album":
		var album models.Album
		if err := c.BodyParser(&album); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}
		media = album
	case "genre":
		var genre models.Genre
		if err := c.BodyParser(&genre); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}
		media = genre
	case "track":
		var track models.Track
		if err := c.BodyParser(&track); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}
		media = track
	default:
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid media type",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := mstor.Add(ctx, media)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add media",
		})
	}

	return c.JSON(media)
}
