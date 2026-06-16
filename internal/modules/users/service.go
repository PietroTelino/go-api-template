package users

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var ErrEmailAlreadyInUse = errors.New("user.emailInUse")

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

func (service *Service) Register(ctx context.Context, input RegisterInput) (*User, error) {
	existingUser, err := service.repository.FindByEmail(ctx, input.Email)
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

	createdUser, err := service.repository.Create(ctx, input.Name, input.Email, string(hashedPassword))
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}
