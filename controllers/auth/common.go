package auth

import (
	"net/mail"
	"regexp"
	"strings"

	"codeberg.org/mjh/LibRate/models"
)

func isEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
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
