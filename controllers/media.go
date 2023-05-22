package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"librerym/models"
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

// PostRating handles the submission of a user's rating for a specific media item
func PostRating(c *fiber.Ctx) error {
	var input models.RatingInput
	rStorage := models.NewRatingStorage()
	err := json.Unmarshal(c.Body(), &input)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	rating := models.Rating{
		UserID:   input.UserID,
		MediaID:  input.MediaID,
		NumStars: input.NumStars,
	}

	err = rStorage.SaveRating(&rating)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save rating",
		})
	}

	return c.JSON(rating)
}

// GetRecommendations returns media recommendations for a user based on collaborative filtering
func GetRecommendations(c *fiber.Ctx) error {
	userID, _ := strconv.Atoi(c.Params("id"))

	recommendedMedia, err := recommendations.GetMemberRecommendations(userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get recommendations",
		})
	}

	return c.JSON(recommendedMedia)
}
