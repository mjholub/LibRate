package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	uuid "github.com/gofrs/uuid/v5"
	validator "github.com/wagslane/go-password-validator"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/db"
	"codeberg.org/mjh/LibRate/models"
)

// Register handles the creation of a new user
func Register(c *fiber.Ctx) error {
	input, err := parseInput("register", c)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	validatedInput, err := input.Validate()
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	member, err := createMember(validatedInput)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	err = saveMember(member)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Registration successful",
	})
}

func checkPasswordEntropy(password string) error {
	if err := validator.Validate(password, 60.0); err != nil {
		return fmt.Errorf("password entropy too low")
	}

	return nil
}

func (r RegisterInput) Validate() (*models.MemberInput, error) {
	if r.Email == "" && r.MemberName == "" {
		return nil, fmt.Errorf("email or membername required")
	}

	if r.Password == "" {
		return nil, fmt.Errorf("password required")
	}

	if r.Password != r.PasswordConfirm {
		return nil, fmt.Errorf("passwords do not match")
	}

	err := checkPasswordEntropy(r.Password)
	if err != nil {
		return nil, err
	}

	return &models.MemberInput{
		Email:      r.Email,
		MemberName: r.MemberName,
		Password:   r.Password,
	}, nil
}

func ValidatePassword() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse the JSON body
		var input struct {
			Password string `json:"password"`
		}
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid JSON",
			})
		}

		// Validate the password
		err := checkPasswordEntropy(input.Password)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Password entropy too low",
			})
		}

		// If the password is valid, return a success response
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Password is strong enough",
		})
	}
}

func createMember(input *models.MemberInput) (*models.Member, error) {
	in := cleanInput(input)

	passhash, err := hashWithArgon(input.Password)
	if err != nil {
		return nil, err
	}

	member := &models.Member{
		UUID:         uuid.Must(uuid.NewV4()).String(),
		PassHash:     passhash,
		MemberName:   in.MemberName,
		Email:        in.Email,
		RegTimestamp: time.Now().Unix(),
	}

	return member, nil
}

func saveMember(member *models.Member) error {
	conf := cfg.LoadConfig().OrElse(cfg.ReadDefaults())
	dbConn, err := db.Connect(&conf)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer dbConn.Close()
	ms := models.NewMemberStorage(dbConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return ms.Save(ctx, member)
}
