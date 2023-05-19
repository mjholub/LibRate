package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"librerym/models"

	"github.com/gofiber/fiber/v2"
)

func Login(c *fiber.Ctx) error {
	var (
		input        models.LoginInput
		lookupTarget string
		cleaned      models.MemberInput
	)
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid login request",
		})
	}
	cleaned.MemberName = input.MemberName
	cleaned.Email = input.Email
	cleaned = *cleanInput(&cleaned)

	memberStorer := models.NewMemberStorer()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	errChan := make(chan error, 1)
	defer close(errChan)
	defer func() {
		if err := context.DeadlineExceeded; err != nil {
			errChan <- fmt.Errorf("context deadline exceeded")
			cancel()
		}
	}()

	if cleaned.Email == "" && cleaned.MemberName != "" {
		lookupTarget = cleaned.MemberName
	} else if cleaned.Email != "" && cleaned.MemberName == "" {
		lookupTarget = cleaned.Email
	} else {
		lookupTarget = cleaned.MemberName
	}

	member, err := memberStorer.Load(ctx, lookupTarget)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}

	if !checkArgonPassword(member.PassHash, input.Password) {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid email or password",
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Logged in successfully",
	})
}
