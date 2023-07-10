package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	uuid "github.com/gofrs/uuid/v5"
	validator "github.com/wagslane/go-password-validator"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/models"
)

// Register handles the creation of a new user
func (a *AuthService) Register(c *fiber.Ctx) error {
	input, err := parseInput("register", c)
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, err.Error())
	}

	validatedInput, err := input.Validate()
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, err.Error())
	}

	member, err := createMember(validatedInput)
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, err.Error())
	}

	err = a.saveMember(member)
	if err != nil {
		return h.Res(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Registration successful",
	})
}

func checkPasswordEntropy(password string) (entropy float64, err error) {
	return validator.GetEntropy(password), validator.Validate(password, 50.0)
}

func (r RegisterInput) Validate() (*models.MemberInput, error) {
	if r.Email == "" && r.MemberName == "" {
		return nil, fmt.Errorf("email or nickname required")
	}

	if r.Password == "" {
		return nil, fmt.Errorf("password required")
	}

	if r.Password != r.PasswordConfirm {
		return nil, fmt.Errorf("passwords do not match")
	}

	_, err := checkPasswordEntropy(r.Password)
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
	const minEntropy = 50.0
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
		entropy, err := checkPasswordEntropy(input.Password)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": fmt.Sprintf("password too weak: want entropy > %f, got %f", minEntropy, entropy),
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
		RegTimestamp: time.Now(),
		Roles:        []uint8{3},
	}

	return member, nil
}

func (a *AuthService) saveMember(member *models.Member) error {
	ms := models.NewMemberStorage(a.db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return ms.Save(ctx, member)
}
