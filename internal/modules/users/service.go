package users

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type SessionRevoker interface {
	RevokeAllByUser(ctx context.Context, userID string) error
}

type AdminPasswordNotifier interface {
	SendAdminPasswordReset(to string, newPassword string) error
}

type Service struct {
	repository       *Repository
	sessionRevoker   SessionRevoker
	passwordNotifier AdminPasswordNotifier
}

func NewService(repository *Repository, sessionRevoker SessionRevoker, passwordNotifier AdminPasswordNotifier) *Service {
	return &Service{repository: repository, sessionRevoker: sessionRevoker, passwordNotifier: passwordNotifier}
}

type RegisterInput struct {
	Name        string
	Email       string
	Password    string
	Preferences map[string]any
}

type CreateInput struct {
	Name        string
	Email       string
	Password    string
	Role        string
	Preferences map[string]any
}

type UpdateInput struct {
	Name  *string
	Email *string
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func passwordMatches(hashedPassword string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

func (service *Service) createUser(ctx context.Context, input CreateInput) (*User, error) {
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

	hashedPassword, err := hashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	return service.repository.Create(ctx, CreateParams{
		Name:        input.Name,
		Email:       email,
		Password:    hashedPassword,
		Role:        input.Role,
		Preferences: input.Preferences,
	})
}

func (service *Service) findExisting(ctx context.Context, id string) (*User, error) {
	if !IsValidID(id) {
		return nil, ErrUserNotFound
	}

	user, err := service.repository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (service *Service) Register(ctx context.Context, input RegisterInput) (*User, error) {
	return service.createUser(ctx, CreateInput{
		Name:        input.Name,
		Email:       input.Email,
		Password:    input.Password,
		Role:        RoleUser,
		Preferences: input.Preferences,
	})
}

func (service *Service) Create(ctx context.Context, actorRole string, input CreateInput) (*User, error) {
	if input.Role == "" {
		input.Role = RoleUser
	}

	if err := ValidateRoleAssignment(actorRole, input.Role); err != nil {
		return nil, err
	}

	return service.createUser(ctx, input)
}

func (service *Service) List(ctx context.Context) ([]*User, error) {
	return service.repository.FindAllActive(ctx)
}

func (service *Service) ListDeleted(ctx context.Context) ([]*User, error) {
	return service.repository.FindAllDeleted(ctx)
}

func (service *Service) GetByID(ctx context.Context, id string) (*User, error) {
	return service.findExisting(ctx, id)
}

func (service *Service) Update(ctx context.Context, id string, input UpdateInput) (*User, error) {
	user, err := service.findExisting(ctx, id)
	if err != nil {
		return nil, err
	}

	if user.Role == RoleGod {
		return nil, ErrCannotModifyGod
	}

	if input.Email != nil {
		email := NormalizeEmail(*input.Email)

		if err := ValidateEmail(email); err != nil {
			return nil, err
		}

		input.Email = &email
	}

	updatedUser, err := service.repository.Update(ctx, id, UpdateParams{Name: input.Name, Email: input.Email})
	if err != nil {
		return nil, err
	}

	if updatedUser == nil {
		return nil, ErrUserNotFound
	}

	return updatedUser, nil
}

func (service *Service) ChangeMyPassword(ctx context.Context, userID string, currentPassword string, newPassword string) error {
	user, err := service.findExisting(ctx, userID)
	if err != nil {
		return err
	}

	if !passwordMatches(user.Password, currentPassword) {
		return ErrCurrentPasswordWrong
	}

	if err := ValidatePassword(newPassword); err != nil {
		return err
	}

	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return err
	}

	if err := service.repository.UpdatePassword(ctx, userID, hashedPassword); err != nil {
		return err
	}

	return service.sessionRevoker.RevokeAllByUser(ctx, userID)
}

func (service *Service) UpdateMyTheme(ctx context.Context, userID string, theme string) error {
	if theme != "light" && theme != "dark" {
		return ErrInvalidTheme
	}

	if _, err := service.findExisting(ctx, userID); err != nil {
		return err
	}

	return service.repository.UpdateTheme(ctx, userID, theme)
}

func (service *Service) SelfDelete(ctx context.Context, userID string, password string) (*User, error) {
	user, err := service.findExisting(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !passwordMatches(user.Password, password) {
		return nil, ErrIncorrectPassword
	}

	if user.DeletedAt != nil {
		return nil, ErrAlreadyDeleted
	}

	if err := service.repository.SoftDelete(ctx, userID); err != nil {
		return nil, err
	}

	return user, service.sessionRevoker.RevokeAllByUser(ctx, userID)
}

func (service *Service) Delete(ctx context.Context, id string) (*User, error) {
	user, err := service.findExisting(ctx, id)
	if err != nil {
		return nil, err
	}

	if user.Role == RoleGod {
		return nil, ErrCannotModifyGod
	}

	if user.DeletedAt != nil {
		return nil, ErrAlreadyDeleted
	}

	if err := service.repository.SoftDelete(ctx, id); err != nil {
		return nil, err
	}

	return user, service.sessionRevoker.RevokeAllByUser(ctx, id)
}

func (service *Service) Restore(ctx context.Context, id string) (*User, error) {
	user, err := service.findExisting(ctx, id)
	if err != nil {
		return nil, err
	}

	if user.DeletedAt == nil {
		return nil, ErrNotDeleted
	}

	return user, service.repository.Restore(ctx, id)
}

func (service *Service) Inactivate(ctx context.Context, id string) (*User, error) {
	user, err := service.findExisting(ctx, id)
	if err != nil {
		return nil, err
	}

	if user.Role == RoleGod {
		return nil, ErrCannotInactivateGod
	}

	if user.InactivatedAt != nil {
		return nil, ErrAlreadyInactive
	}

	if err := service.repository.Inactivate(ctx, id); err != nil {
		return nil, err
	}

	return user, service.sessionRevoker.RevokeAllByUser(ctx, id)
}

func (service *Service) Reactivate(ctx context.Context, id string) (*User, error) {
	user, err := service.findExisting(ctx, id)
	if err != nil {
		return nil, err
	}

	if user.InactivatedAt == nil {
		return nil, ErrAlreadyActive
	}

	return user, service.repository.Reactivate(ctx, id)
}

func (service *Service) AdminResetPassword(ctx context.Context, id string) (*User, error) {
	user, err := service.findExisting(ctx, id)
	if err != nil {
		return nil, err
	}

	if user.Role == RoleGod {
		return nil, ErrCannotResetGodPassword
	}

	newPassword, err := generateRandomPassword()
	if err != nil {
		return nil, err
	}

	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return nil, err
	}

	if err := service.repository.UpdatePassword(ctx, id, hashedPassword); err != nil {
		return nil, err
	}

	if err := service.sessionRevoker.RevokeAllByUser(ctx, id); err != nil {
		return nil, err
	}

	if err := service.passwordNotifier.SendAdminPasswordReset(user.Email, newPassword); err != nil {
		return nil, fmt.Errorf("senha alterada, mas o e-mail não foi enviado: %w", err)
	}

	return user, nil
}

func (service *Service) HardDelete(ctx context.Context, id string) (*User, error) {
	user, err := service.findExisting(ctx, id)
	if err != nil {
		return nil, err
	}

	if user.Role == RoleGod {
		return nil, ErrCannotDeleteGod
	}

	return user, service.repository.HardDelete(ctx, id)
}
