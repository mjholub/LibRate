package cfg

import (
	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	FailedField string
	Tag         string
	Value       interface{}
}

func Validate(conf *Config, validate *validator.Validate) (errs []ValidationError) {
	errors := validate.Struct(conf)
	if errors != nil {
		for _, err := range errors.(validator.ValidationErrors) {
			var ve ValidationError
			ve.FailedField = err.Field()
			ve.Tag = err.Tag()
			ve.Value = err.Value()

			errs = append(errs, ve)
		}
	}
	return errs
}
