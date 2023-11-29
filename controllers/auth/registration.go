package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	uuid "github.com/gofrs/uuid/v5"

	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/lib/redist"
	"codeberg.org/mjh/LibRate/models/member"
)

// Register handles the creation of a new user
func (a *Service) Register(c *fiber.Ctx) error {
	a.log.Debug().Msg("Registration request")
	input, err := parseRegistrationInput(c)
	if err != nil {
		return h.Res(c, fiber.StatusBadRequest, err.Error())
	}
	if input == nil {
		return h.Res(c, fiber.StatusInternalServerError, "Cannot parse input")
	}
	a.log.Debug().Msg("Parsed input")

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

func (r RegisterInput) Validate() (*member.Input, error) {
	if r.Email == "" && r.MemberName == "" {
		return nil, fmt.Errorf("email or nickname required")
	}

	password := string(r.Password)

	if password == "" {
		return nil, fmt.Errorf("password required")
	}

	if password != string(r.PasswordConfirm) {
		return nil, fmt.Errorf("passwords do not match")
	}

	_, err := redist.CheckPasswordEntropy(password)
	if err != nil {
		return nil, err
	}

	return &member.Input{
		Email:      r.Email,
		MemberName: r.MemberName,
		Password:   r.Password,
	}, nil
}

func createMember(input *member.Input) (*member.Member, error) {
	in := cleanInput(input)

	passhash, err := hashWithArgon([]byte(input.Password))
	if err != nil {
		return nil, err
	}

	member := &member.Member{
		UUID:         uuid.Must(uuid.NewV4()),
		PassHash:     passhash,
		MemberName:   in.MemberName,
		Email:        in.Email,
		RegTimestamp: time.Now(),
		Roles:        []string{"member"},
	}

	return member, nil
}

func (a *Service) saveMember(member *member.Member) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return a.ms.Save(ctx, member)
}
