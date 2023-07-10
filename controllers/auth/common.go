package auth

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"

	"codeberg.org/mjh/LibRate/cfg"
	"codeberg.org/mjh/LibRate/models"
)

type (
	// RegisterInput is the input for the registration request
	RegisterInput struct {
		Email           string `json:"email"`
		MemberName      string `json:"membername"`
		Password        string `json:"password"`
		PasswordConfirm string `json:"passwordConfirm"`
	}

	// LoginInput is the input for the login request
	LoginInput struct {
		Email      string `json:"email"`
		MemberName string `json:"membername"`
		Password   string `json:"password"`
	}

	// AuthService allows dependency injection for the controller methods,
	// so that the db connection needn't be created in the controller methods
	AuthService struct {
		conf *cfg.Config
		db   *sqlx.DB
	}

	// RegLoginInput is an union (feature introduced in Go 1.18) of RegisterInput and LoginInput
	RegLoginInput interface {
		RegisterInput | LoginInput
	}

	Validator interface {
		Validate() (*models.MemberInput, error)
	}
)

// NewAuthService creates an instance of the AuthService struct
// and returns a pointer to it
// It should be used within the routes package
// where the db connection and config are passed from the main package
func NewAuthService(conf *cfg.Config, db *sqlx.DB) *AuthService {
	return &AuthService{conf, db}
}

func isEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

//  parseInput parses the input from the request body to be used in the controller
func parseInput(reqType string, c *fiber.Ctx) (Validator, error) {
	switch reqType {
	case "register":
		var input RegisterInput
		err := c.BodyParser(&input)
		if err != nil {
			return input, fmt.Errorf("invalid registration request")
		}

		return input, nil
	case "login":
		var input LoginInput
		if err := c.BodyParser(&input); err != nil {
			return nil, err
		}
		return input, nil
	}
	return nil, fmt.Errorf("unknown request type")
}

// cleanRegInput cleans the input from non-ASCII and unsafe characters
func cleanInput(input *models.MemberInput) *models.MemberInput {
	input.MemberName = strings.TrimSpace(input.MemberName)
	re := regexp.MustCompile("[^a-zA-Z0-9]+")
	mailRe := regexp.MustCompile("[^a-zA-Z0-9@.]+")
	input.MemberName = re.ReplaceAllString(input.MemberName, "")
	input.Email = strings.TrimSpace(input.Email)
	input.Email = strings.ToLower(input.Email)
	input.Email = mailRe.ReplaceAllString(input.Email, "")
	return input
}
