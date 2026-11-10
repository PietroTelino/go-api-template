package users

import "errors"

var (
	ErrEmailAlreadyInUse     = errors.New("user.emailInUse")
	ErrEmailRequired         = errors.New("email.required")
	ErrEmailInvalidFormat    = errors.New("email.invalidFormat")
	ErrPasswordTooShort      = errors.New("password.minLength")
	ErrPasswordTooLong       = errors.New("password.maxLength")
	ErrPasswordNoSpecialChar = errors.New("password.specialChar")

	ErrUserNotFound         = errors.New("user.notFound")
	ErrAlreadyDeleted       = errors.New("user.alreadyDeleted")
	ErrNotDeleted           = errors.New("user.notDeleted")
	ErrAlreadyInactive      = errors.New("user.alreadyInactive")
	ErrAlreadyActive        = errors.New("user.alreadyActive")
	ErrIncorrectPassword    = errors.New("user.incorrectPassword")
	ErrCurrentPasswordWrong = errors.New("user.currentPasswordWrong")
	ErrInvalidTheme         = errors.New("user.invalidTheme")
	ErrInvalidRole          = errors.New("user.invalidRole")

	ErrCannotAssignRole    = errors.New("user.cannotAssignRole")
	ErrCannotModifyGod     = errors.New("user.cannotModifyGod")
	ErrCannotInactivateGod = errors.New("user.cannotInactivateGod")
	ErrCannotDeleteGod     = errors.New("user.cannotDeleteGod")
)
