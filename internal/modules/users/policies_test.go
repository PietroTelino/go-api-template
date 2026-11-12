package users

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizeEmail(t *testing.T) {
	if got := NormalizeEmail("  Maria@Email.COM  "); got != "maria@email.com" {
		t.Errorf("NormalizeEmail = %q, esperado %q", got, "maria@email.com")
	}
}

func TestValidateEmail(t *testing.T) {
	testCases := []struct {
		email string
		want  error
	}{
		{"maria@email.com", nil},
		{"a@b.co", nil},
		{"", ErrEmailRequired},
		{"semarroba", ErrEmailInvalidFormat},
		{"sem@ponto", ErrEmailInvalidFormat},
		{"com espaco@email.com", ErrEmailInvalidFormat},
		{"duas@@email.com", ErrEmailInvalidFormat},
	}

	for _, testCase := range testCases {
		t.Run(testCase.email, func(t *testing.T) {
			if got := ValidateEmail(testCase.email); !errors.Is(got, testCase.want) {
				t.Errorf("ValidateEmail(%q) = %v, esperado %v", testCase.email, got, testCase.want)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		want     error
	}{
		{"válida", "Senha@123", nil},
		{"curta", "a@1", ErrPasswordTooShort},
		{"sem caractere especial", "12345678", ErrPasswordNoSpecialChar},
		{"8 caracteres com acento", "ééééééé@", nil},
		{"72 bytes", strings.Repeat("a", 71) + "@", nil},
		{"73 bytes", strings.Repeat("a", 72) + "@", ErrPasswordTooLong},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := ValidatePassword(testCase.password); !errors.Is(got, testCase.want) {
				t.Errorf("ValidatePassword(%q) = %v, esperado %v", testCase.password, got, testCase.want)
			}
		})
	}
}

func TestValidateRoleAssignment(t *testing.T) {
	testCases := []struct {
		actorRole  string
		targetRole string
		want       error
	}{
		{RoleAdministrator, RoleUser, nil},
		{RoleGod, RoleUser, nil},
		{RoleGod, RoleAdministrator, nil},
		{RoleAdministrator, RoleAdministrator, ErrCannotAssignRole},
		{RoleGod, RoleGod, ErrCannotAssignRole},
		{RoleAdministrator, RoleGod, ErrCannotAssignRole},
		{RoleGod, "chefe", ErrInvalidRole},
	}

	for _, testCase := range testCases {
		t.Run(testCase.actorRole+" cria "+testCase.targetRole, func(t *testing.T) {
			if got := ValidateRoleAssignment(testCase.actorRole, testCase.targetRole); !errors.Is(got, testCase.want) {
				t.Errorf("ValidateRoleAssignment(%q, %q) = %v, esperado %v", testCase.actorRole, testCase.targetRole, got, testCase.want)
			}
		})
	}
}
