package controllers

import (
	"context"
	"encoding/json"
	"fmt"
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
		GetMediaReviews(c *fiber.Ctx) error
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
func (rc *ReviewController) GetMediaReviews(c *fiber.Ctx) error {
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

// GetByID retrieves a single review by its ID
func (rc *ReviewController) GetByID(c *fiber.Ctx) error {
	reviewID := c.Params("id")
	id, err := strconv.ParseInt(reviewID, 10, 64)
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, "Invalid review ID \""+reviewID+"\"")
	}
	review, err := rc.rs.Get(c.UserContext(), id)
	if err != nil {
		rc.ms.Log.Error().Err(err).Msgf(err.Error())
		return h.Res(c, fiber.StatusInternalServerError, "Failed to fetch review")
	}
	return c.JSON(review)
}

// GetLatestRatings retrieves the latest reviews for a specific media item based on the media ID
func (rc *ReviewController) GetLatest(c *fiber.Ctx) error {
	// Extract limit and offset parameters from the query string.
	limit, err := strconv.Atoi(c.Query("limit", "5"))
	if err != nil || limit < 1 || limit > 100 {
		return h.Res(c, fiber.StatusBadRequest, "Invalid limit")
	}
	// offset is the number of ratings to skip
	offset, err := strconv.Atoi(c.Query("offset", "0"))
	if err != nil || offset < 0 || offset > 50 {
		return h.Res(c, fiber.StatusBadRequest, "Invalid offset")
	}

	// Call the GetLatest function with the provided limit and offset.
	ratings, err := rc.rs.GetLatest(c.Context(), limit, offset)
	if err != nil {
		return h.Res(c, fiber.StatusNotFound, err.Error())
	}

	// Return the ratings as a JSON response.
	return c.JSON(ratings)
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
		average, err := rc.getTrackAverageScore(c.UserContext(), mediaID)
		if err != nil {
			return h.Res(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(average)
	case "album":
		average, err := rc.getAlbumAverageScore(c.UserContext(), mediaID)
		if err != nil {
			return h.Res(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(average)
	default:
		return h.Res(c, fiber.StatusNotImplemented,
			fmt.Sprintf(`Fetching average score for this media type (%s) is not implemented yet.
			Feel free to open an issue on codeberg or github`, mediaKind))
	}
}

func (rc *ReviewController) getTrackAverageScore(
	ctx context.Context, id uuid.UUID,
) (*models.RatingAverage, error) {
	average, err := rc.rs.GetAverageStars(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get average score for track with ID %s: %v", id.String(), err.Error())
	}
	return &models.RatingAverage{
		BaseRatingScore:         average,
		SecondaryRatingTypes:    nil,
		SecondaryRatingAverages: nil,
	}, nil
}

func (rc *ReviewController) getAlbumAverageScore(
	ctx context.Context, id uuid.UUID,
) (*models.RatingAverage, error) {
	// get track IDs for the given album
	trackIDs, err := rc.ms.GetAlbumTrackIDs(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch track IDs for album %s: %w", id.String(), err)
	}
	trackAverages := make([]models.SecondaryRatingAverage, len(trackIDs))
	var trackScore float64
	for i := range trackIDs {
		// fetch the average rating for a single track
		trackScore, err = rc.rs.GetAverageStars(ctx, trackIDs[i])
		if err != nil {
			return nil, fmt.Errorf(
				"failed to fetch average track rating when trying to retrieve track ratings for album with ID %s: %w",
				id.String(), err)
		}
		trackAverages = append(trackAverages, models.SecondaryRatingAverage{
			MediaID:   trackIDs[i],
			MediaKind: "track",
			Score:     trackScore,
		})
	}
	albumAverage, err := rc.rs.GetAverageStars(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting average score for album with ID %s: %w", id.String(), err)
	}
	return &models.RatingAverage{
		BaseRatingScore:         albumAverage,
		SecondaryRatingTypes:    &[]string{"track"},
		SecondaryRatingAverages: trackAverages,
	}, nil
}
