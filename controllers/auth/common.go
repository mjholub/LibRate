package auth

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"

	"codeberg.org/mjh/LibRate/models"
)

type RegisterInput struct {
	Email           string `json:"email"`
	MemberName      string `json:"membername"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"passwordConfirm"`
}

type LoginInput struct {
	Email      string `json:"email"`
	MemberName string `json:"membername"`
	Password   string `json:"password"`
}

type RegLoginInput interface {
	RegisterInput | LoginInput
}

type Validator interface {
	Validate() (*models.MemberInput, error)
}

func isEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

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
