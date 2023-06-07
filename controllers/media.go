package controllers

import (
	"context"
	"net/http"
	"reflect"
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
	var media any

	if err := c.BodyParser(&media); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	value := media.(map[string]interface{})
	// unmarshal the JSON payload into a struct
	switch value["type"].(string) {
	case "film":
		film := models.Film{
			Title: value["title"].(string),
			Year:  value["year"].(int),
			Cast: models.Cast{
				Actors:    value["actors"].([]models.Person),
				Directors: value["directors"].([]models.Person),
			},
		}
		media = film
	case "album":
		album := models.Album{
			Name:        value["name"].(string),
			Artists:     value["artists"].([]models.Person),
			ReleaseDate: value["releaseDate"].(time.Time),
			Genres:      value["genres"].([]string),
			Keywords:    value["keywords"].([]string),
			Duration:    value["duration"].(time.Duration),
			Tracks:      value["tracks"].([]models.Track),
		}
		media = album
	case "genre":
		genre := models.Genre{
			Name:        value["name"].(string),
			Description: value["description"].(string),
			Keywords:    value["keywords"].([]string),
		}
		media = genre
	case "track":
		track := models.Track{
			Name:     value["name"].(string),
			Artists:  value["artists"].([]models.Person),
			Duration: value["duration"].(time.Duration),
			Lyrics:   value["lyrics"].(string),
		}
		media = track
	default:
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid media type",
		})
	}

	err := mstor.Add(ctx, &media, reflect.TypeOf(media))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add media",
		})
	}

	return c.JSON(media)
}
