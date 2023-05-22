package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	uuid "github.com/gofrs/uuid/v5"
	validator "github.com/wagslane/go-password-validator"

	"codeberg.org/mjh/LibRate/models"
)

// Register handles the creation of a new user
func Register(c *fiber.Ctx) error {
	var (
		input   models.RegisterInput
		inClean models.MemberInput
	)
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid registration request",
		})
	}

	if input.Password != input.PasswordConfirm {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Passwords do not match",
		})
	}

	if err := validator.Validate(input.Password, 50.0); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	inClean.MemberName = input.MemberName
	inClean.Email = input.Email

	inClean = *cleanInput(&inClean)

	if !isEmail(inClean.Email) {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid e-mail format",
		})
	}

	if len(inClean.MemberName) > 20 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Member name too long",
		})
	}

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

	passhash, err := hashWithArgon(input.Password)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	member := &models.Member{
		UUID:         uuid.Must(uuid.NewV4()).String(),
		PassHash:     passhash,
		MemberName:   inClean.MemberName,
		Email:        inClean.Email,
		RegTimestamp: time.Now().Unix(),
	}

	err = memberStorer.Save(ctx, member)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Registration successful",
	})
}
