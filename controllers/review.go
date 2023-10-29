package controllers

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid/v5"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/models"
)

type (
	// IReviewController is the interface for the review controller
	// It defines the methods that the review controller must implement
	// This is useful for mocking the review controller in unit tests
	IReviewController interface {
		GetRatings(c *fiber.Ctx) error
		GetLatestRatings(c *fiber.Ctx) error
		GetAverageRating(c *fiber.Ctx) error
		PostRating(c *fiber.Ctx) error
	}

	// ReviewController is the controller for review endpoints
	ReviewController struct {
		rs *models.RatingStorage
		ms *models.MediaStorage
	}
)

func NewReviewController(rs models.RatingStorage) *ReviewController {
	return &ReviewController{rs: &rs}
}

// GetMediaRatings retrieves reviews for a specific media item based on the media ID
func (rc *ReviewController) GetMediaRatings(c *fiber.Ctx) error {
	mediaID, err := uuid.FromString(c.Params("media_id"))
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, "Invalid media ID")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	reviews, err := rc.rs.GetByMediaID(ctx, mediaID)
	if err != nil {
		return h.Res(c, fiber.StatusNotFound, "Ratings not found")
	}

	return c.JSON(reviews)
}

// GetLatestRatings retrieves the latest reviews for a specific media item based on the media ID
func (rc *ReviewController) GetLatestRatings(ctx *fiber.Ctx) error {
	// Extract limit and offset parameters from the query string.
	limit, err := strconv.Atoi(ctx.Query("limit", "5"))
	if err != nil {
		return h.Res(ctx, fiber.StatusBadRequest, "Invalid limit")
	}
	offset, err := strconv.Atoi(ctx.Query("offset", "0"))
	if err != nil {
		return h.Res(ctx, fiber.StatusBadRequest, "Invalid offset")
	}

	// Call the GetLatest function with the provided limit and offset.
	ratings, err := rc.rs.GetLatest(ctx.Context(), limit, offset)
	if err != nil {
		return h.Res(ctx, fiber.StatusNotFound, err.Error())
	}

	// Return the ratings as a JSON response.
	return ctx.JSON(ratings)
}

// GetAverageRatings retrieves the average number of stars for the general models.Rating type
// (i.e. not track or cast ratings)
func (rc *ReviewController) GetAverageRatings(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mediaID, err := uuid.FromString(c.Params("id"))
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, "Invalid media ID")
	}
	avgStars, err := rc.rs.GetAverageStars(ctx, &models.Rating{}, mediaID)
	if err != nil {
		return h.Res(c, fiber.StatusNotFound, "Failed to fetch average stars")
	}

	return c.JSON(avgStars)
}

// PostRating handles the submission of a user's review for a specific media item
func (rc *ReviewController) PostRating(c *fiber.Ctx) error {
	var input models.RatingInput
	err := json.Unmarshal(c.Body(), &input)
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, "Invalid input")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = rc.rs.New(ctx, &input)
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, "Failed to add rating")
	}

	return c.JSON(fiber.Map{
		"message": "Rating added successfully",
	})
}

// UpdateRating handles the update of a user's review for a specific media item
// The types of values that can be updated are defined by the union type models.UpdateableKeyTypes
// This way, things like the date of the review are not updateable
func (rc *ReviewController) UpdateRating(c *fiber.Ctx) error {
	ratingID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, "Invalid rating ID")
	}

	var keysToUpdate []string
	err = json.Unmarshal(c.Body(), &keysToUpdate)
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, "Invalid input")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// WARN: not sure if this is correct, but one cannot have type params on receiver methods
	err = models.UpdateRating(ctx, rc.rs, ratingID, keysToUpdate)
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, "Failed to update rating")
	}

	return c.JSON(fiber.Map{
		"message": "Rating updated successfully",
	})
}

// DeleteRating handles the deletion of a user's review for a specific media item
func (rc *ReviewController) DeleteRating(c *fiber.Ctx) error {
	ratingID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, "Invalid rating ID")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = rc.rs.Delete(ctx, ratingID)
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, "Failed to delete rating")
	}

	return c.JSON(fiber.Map{
		"message": "Rating deleted successfully",
	})
}

// GetAverageRating fetches the average (float64) rating ("stars") score based on a given media UUID, kind
// and rating type
func (rc *ReviewController) GetAverageRating(c *fiber.Ctx) error {
	mediaID, err := uuid.FromString(c.Params("media_id"))
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, "Invalid media ID")
	}

	mediaKind, err := rc.ms.GetKind(c.UserContext(), mediaID) 
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, "Failed to fetch media kind")
	}
	
	switch mediaKind {
	case "track":
	average, err := rc.rs.GetAverageStars(c.UserContext(), mediaID)
	if err != nil {
		return h.Res(c, fiber.StatusNotFound, "Failed to fetch average stars")
	}
}
