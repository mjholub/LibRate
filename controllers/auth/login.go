package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/models"

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

	if cleaned.Email == "" && cleaned.MemberName != "" {
		lookupTarget = cleaned.MemberName
	} else if cleaned.Email != "" && cleaned.MemberName == "" {
		lookupTarget = cleaned.Email
	} else {
		lookupTarget = cleaned.MemberName
	}

	conf := cfg.LoadDgraph()

	ms, conn, err := models.NewMemberStorage(*conf)
	defer conn.Close()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to connect to database",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	statusChan := make(chan error, 1)

	go func() {
		defer close(statusChan)

		member, err := ms.Load(ctx, lookupTarget)
		if err != nil {
			statusChan <- err
			return
		}

		if !checkArgonPassword(member.PassHash, input.Password) {
			statusChan <- errors.New("invalid email or password")
			return
		}
	}()

	select {
	case <-ctx.Done():
		return c.Status(http.StatusRequestTimeout).JSON(fiber.Map{
			"message": "Request timeout",
		})
	case err := <-statusChan:
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"message": err.Error(),
			})
		}
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Logged in successfully",
	})
}
