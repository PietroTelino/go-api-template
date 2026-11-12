package users

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/PietroTelino/go-api-template/internal/validation"
)

const (
	minPasswordLength = 8
	maxPasswordBytes  = 72
)

var (
	emailPattern       = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	specialCharPattern = regexp.MustCompile(`[^a-zA-Z0-9]`)
)

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func ValidateEmail(email string) error {
	if email == "" {
		return ErrEmailRequired
	}

	if !emailPattern.MatchString(email) {
		return ErrEmailInvalidFormat
	}

	return nil
}

func ValidatePassword(password string) error {
	if utf8.RuneCountInString(password) < minPasswordLength {
		return ErrPasswordTooShort
	}

	if len(password) > maxPasswordBytes {
		return ErrPasswordTooLong
	}

	if !specialCharPattern.MatchString(password) {
		return ErrPasswordNoSpecialChar
	}

	return nil
}

func IsValidID(id string) bool {
	return validation.IsUUID(id)
}

func ValidateRoleAssignment(actorRole string, targetRole string) error {
	switch targetRole {
	case RoleUser:
		return nil
	case RoleAdministrator:
		if actorRole == RoleGod {
			return nil
		}
		return ErrCannotAssignRole
	case RoleGod:
		return ErrCannotAssignRole
	default:
		return ErrInvalidRole
	}
}
