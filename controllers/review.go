package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"

	"codeberg.org/mjh/LibRate/models"
)

// GetRatings retrieves reviews for a specific media item based on the media ID
func GetRatings(c *fiber.Ctx) error {
	rStorage := models.NewRatingStorage()

	ratingID, err := uuid.FromString(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid media ID",
		})
	}

	reviews, err := rStorage.GetByMediaID(context.Background(), ratingID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Ratings not found",
		})
	}

	return c.JSON(reviews)
}

// GetAverageRatings retrieves the average number of stars for the general models.Rating type
// (i.e. not track or cast ratings)
func GetAverageRatings(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mediaID, err := uuid.FromString(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid media ID",
		})
	}
	avgStars, err := models.GetAverageStars(ctx, &models.Rating{}, mediaID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Ratings not found",
		})
	}

	return c.JSON(avgStars)
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = rs.New(ctx, &input)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add rating",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Rating added successfully",
	})
}

func UpdateRating(c *fiber.Ctx) error {
	ratingID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid rating ID",
		})
	}

	var keysToUpdate []string
	err = json.Unmarshal(c.Body(), &keysToUpdate)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = models.UpdateRating(ctx, ratingID, keysToUpdate)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update rating",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Rating updated successfully",
	})
}
