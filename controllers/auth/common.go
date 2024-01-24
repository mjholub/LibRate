package auth

import (
	"net/mail"
	"regexp"
	"strconv"
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
		Email       string `json:"email,omitempty"`
		MemberName  string `json:"membername,omitempty"`
		Password    string `json:"password"`
		SessionTime int32  `json:"session_time" default:"30"` // in minutes. Setting to 2^31-1 is used to keep user signed in
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

func parseLoginInput(c *fiber.Ctx, log *zerolog.Logger) (*LoginInput, error) {
	var input LoginInput
	var err error
	if input.Email != "" || input.MemberName != "" {
		if !isEmail(input.Email) {
			return nil, h.Res(c, fiber.StatusBadRequest, "Invalid email address")
		}
	}
	input.Email = c.FormValue("email", "")
	input.MemberName = c.FormValue("membername", "")
	input.Password = c.FormValue("password", "")
	sessionTime := c.FormValue("session_time", "30")
	sTimeout, err := strconv.Atoi(sessionTime)
	if err != nil {
		log.Log().Err(err).Msgf("Failed to parse session time %s for member %s (%s): %s",
			sessionTime, input.MemberName, input.Email, err.Error())
		input.SessionTime = 30
	} else {
		if sTimeout < 0 || sTimeout > 2147483647 {
			input.SessionTime = 2147483647 // assume the user used -1 as infinite session time, also protects from overflow
		} else {
			// nolint:gosec //check for overflow is done above
			input.SessionTime = int32(sTimeout)
		}
	}
	if err != nil {
		return nil, h.Res(c, fiber.StatusBadRequest, "Invalid login request")
	}
	return &input, nil
}

func parseRegistrationInput(c *fiber.Ctx) (input *RegisterInput, err error) {
	if form, err := c.MultipartForm(); err == nil {
		if form.Value["password"][0] != form.Value["passwordConfirm"][0] {
			return nil, h.Res(c, fiber.StatusBadRequest, "Passwords do not match")
		}
		if form.Value["email"] == nil || form.Value["membername"] == nil {
			return nil, h.Res(c, fiber.StatusBadRequest, "Email and nickname required")
		}
		if !isEmail(form.Value["email"][0]) {
			return nil, h.Res(c, fiber.StatusBadRequest, "Invalid email address")
		}
		input = &RegisterInput{
			Email:           form.Value["email"][0],
			MemberName:      form.Value["membername"][0],
			Password:        form.Value["password"][0],
			PasswordConfirm: form.Value["passwordConfirm"][0],
			Roles:           []string{"member"},
		}
		return input, nil
	}

	return nil, h.Res(c, fiber.StatusBadRequest, "Invalid registration request")
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
