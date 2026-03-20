package model

import "github.com/go-playground/validator/v10"

type CustomValidator struct {
	Validator *validator.Validate
}

func (v *CustomValidator) Validate(i any) error {
	return v.Validator.Struct(i)
}
