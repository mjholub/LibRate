package controllers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"codeberg.org/mjh/LibRate/internal/client"
	h "codeberg.org/mjh/LibRate/internal/handlers"
	services "codeberg.org/mjh/LibRate/recommendation/go/services"
)

// GetRecommendations returns media recommendations for a user based on collaborative filtering
// TODO: the actual underlying functionality, i.e. the recommendations server
func GetRecommendations(c *fiber.Ctx) error {
	mID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		//nolint:errcheck
		return h.Res(c, fiber.StatusBadRequest, fmt.Sprintf("Invalid member ID %s (must be an integer)", c.Params("id")))
	}

	memberID := int32(mID)

	conn, err := client.ConnectToService(context.Background(), "recommendation", "50051")
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, "Failed to connect to recommendation service")
	}
	defer conn.Close()

	s := services.NewRecommendationServiceClient(conn)

	recommendedMedia, err := s.GetRecommendations(context.Background(), &services.GetRecommendationsRequest{
		MemberId: memberID,
	})
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, "Failed to get recommendations")
	}

	return c.JSON(recommendedMedia)
}
