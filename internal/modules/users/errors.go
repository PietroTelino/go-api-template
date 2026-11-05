package users

import "errors"

var (
	ErrEmailAlreadyInUse     = errors.New("user.emailInUse")
	ErrEmailRequired         = errors.New("email.required")
	ErrEmailInvalidFormat    = errors.New("email.invalidFormat")
	ErrPasswordTooShort      = errors.New("password.minLength")
	ErrPasswordTooLong       = errors.New("password.maxLength")
	ErrPasswordNoSpecialChar = errors.New("password.specialChar")
)
