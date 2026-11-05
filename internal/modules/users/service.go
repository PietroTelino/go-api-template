package users

import (
	"context"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

type RegisterInput struct {
	Name        string
	Email       string
	Password    string
	Preferences map[string]any
}

func (service *Service) Register(ctx context.Context, input RegisterInput) (*User, error) {
	email := NormalizeEmail(input.Email)

	if err := ValidateEmail(email); err != nil {
		return nil, err
	}

	if err := ValidatePassword(input.Password); err != nil {
		return nil, err
	}

	existingUser, err := service.repository.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		return nil, ErrEmailAlreadyInUse
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return service.repository.Create(ctx, CreateParams{
		Name:        input.Name,
		Email:       email,
		Password:    string(hashedPassword),
		Role:        RoleUser,
		Preferences: input.Preferences,
	})
}
