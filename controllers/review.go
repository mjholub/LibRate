package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"codeberg.org/mjh/LibRate/models"
)

// GetReviews retrieves reviews for a specific media item based on the media ID
func GetReviews(c *fiber.Ctx) error {
	mediaID, _ := strconv.Atoi(c.Params("id"))

	reviews, err := models.GetReviewsByMediaID(mediaID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Reviews not found",
		})
	}

	return c.JSON(reviews)
}

// PostReview handles the submission of a user's review for a specific media item
func PostReview(c *fiber.Ctx) error {
	var input models.ReviewInput
	err := json.Unmarshal(c.Body(), &input)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	review := models.Review{
		UserID:     input.UserID,
		MediaID:    input.MediaID,
		ReviewText: input.ReviewText,
	}

	err = models.SaveReview(&review)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save review",
		})
	}

	return c.JSON(review)
}

// GetPinnedReviews returns pinned reviews for a user profile
func GetPinnedReviews(c *fiber.Ctx) error {
	userID, _ := strconv.Atoi(c.Params("id"))

	pinnedReviews, err := models.GetPinnedReviewsByUserID(userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get pinned reviews",
		})
	}

	return c.JSON(pinnedReviews)
}
