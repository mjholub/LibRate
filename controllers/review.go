package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v2"

	"codeberg.org/mjh/LibRate/models"
)

// GetRatings retrieves reviews for a specific media item based on the media ID
func GetRatings(c *fiber.Ctx) error {
	rStorage := models.NewRatingStorage()

	reviews, err := rStorage.Get(context.Background(), c.Params("id"))
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Ratings not found",
		})
	}

	return c.JSON(reviews)
}

// PostRating handles the submission of a user's review for a specific media item
func PostRating(c *fiber.Ctx) error {
	var input models.RatingInput
	rs := models.NewRatingStorage()
	err := json.Unmarshal(c.Body(), &input)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	review := models.Rating{
		UserID:  input.UserID,
		MediaID: input.MediaID,
		Comment: input.Comment,
	}

	err = rs.SaveRating(&review)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save review",
		})
	}

	return c.JSON(review)
}

// GetPinnedRatings returns pinned reviews for a user profile
func GetPinnedRatings(c *fiber.Ctx) error {
	rs := models.NewRatingStorage()
	pinnedRatings, err := rs.GetPinned(context.TODO())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get pinned reviews",
		})
	}

	return c.JSON(pinnedRatings)
}
