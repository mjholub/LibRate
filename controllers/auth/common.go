package auth

import (
	"net/mail"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/rs/zerolog"

	"codeberg.org/mjh/LibRate/cfg"
	h "codeberg.org/mjh/LibRate/internal/handlers"
	"codeberg.org/mjh/LibRate/models/member"
)

type (
	// RegisterInput is the input for the registration request
	RegisterInput struct {
		Email      string `json:"email"`
		MemberName string `json:"membername"`
		// Password is first temporarily encrypted using RSA and then hashed using argon2id
		// For more details see the internal/crypt package
		Password        string   `json:"password"`
		PasswordConfirm string   `json:"passwordConfirm"`
		Roles           []string `json:"roles"`
	}

	// LoginInput is the input for the login request
	LoginInput struct {
		Email      string `json:"email,omitempty"`
		MemberName string `json:"membername,omitempty"`
		Password   string `json:"password"`
	}

	// Service allows dependency injection for the controller methods,
	// so that the db connection needn't be created in the controller methods
	Service struct {
		conf *cfg.Config
		log  *zerolog.Logger
		ms   member.MemberStorer
		sess *session.Store
	}

	// RegLoginInput is an union (feature introduced in Go 1.18) of RegisterInput and LoginInput
	RegLoginInput interface {
		RegisterInput | LoginInput
	}

	Validator interface {
		Validate() (*member.Input, error)
	}
)

// NewService creates an instance of the Service struct
// and returns a pointer to it
// It should be used within the routes package
// where the db connection and config are passed from the main package
func NewService(
	conf *cfg.Config,
	ms member.MemberStorer,
	log *zerolog.Logger,
	sess *session.Store,
) *Service {
	return &Service{conf, log, ms, sess}
}

func isEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func parseLoginInput(c *fiber.Ctx) (*LoginInput, error) {
	var input LoginInput
	if input.Email != "" || input.MemberName != "" {
		if !isEmail(input.Email) {
			return nil, h.Res(c, fiber.StatusBadRequest, "Invalid email address")
		}
	}
	err := c.BodyParser(&input)
	if err != nil {
		return nil, h.Res(c, fiber.StatusBadRequest, "Invalid login request")
	}
	return &input, nil
}

func parseRegistrationInput(c *fiber.Ctx) (input *RegisterInput, err error) {
	err = c.BodyParser(&input)
	if err != nil {
		return nil, h.Res(c, fiber.StatusBadRequest, "Invalid registration request")
	}
	if input.Password != input.PasswordConfirm {
		return nil, h.Res(c, fiber.StatusBadRequest, "Passwords do not match")
	}
	if input.Email == "" && input.MemberName == "" {
		return nil, h.Res(c, fiber.StatusBadRequest, "Email and nickname required")
	}
	if !isEmail(input.Email) {
		return nil, h.Res(c, fiber.StatusBadRequest, "Invalid email address")
	}
	return input, nil
}

// cleanRegInput cleans the input from non-ASCII and unsafe characters
func cleanInput(input *member.Input) *member.Input {
	input.MemberName = strings.TrimSpace(input.MemberName)
	re := regexp.MustCompile("[^a-zA-Z0-9]+")
	mailRe := regexp.MustCompile("[^a-zA-Z0-9@.]+")
	input.MemberName = re.ReplaceAllString(input.MemberName, "")
	input.Email = strings.TrimSpace(input.Email)
	input.Email = strings.ToLower(input.Email)
	input.Email = mailRe.ReplaceAllString(input.Email, "")
	return input
}
