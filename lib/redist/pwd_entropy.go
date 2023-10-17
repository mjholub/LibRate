package redist

import validator "github.com/wagslane/go-password-validator"

func CheckPasswordEntropy(password string) (entropy float64, err error) {
	return validator.GetEntropy(password), validator.Validate(password, 50.0)
}
