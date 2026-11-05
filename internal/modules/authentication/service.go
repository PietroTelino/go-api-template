package authentication

import (
	"context"
	"errors"
	"time"

	"github.com/PietroTelino/go-api-template/internal/auth"
	"github.com/PietroTelino/go-api-template/internal/modules/sessions"
	"github.com/PietroTelino/go-api-template/internal/modules/users"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	usersRepository    *users.Repository
	sessionsRepository *sessions.Repository
	tokenManager       *auth.TokenManager
}

func NewService(
	usersRepository *users.Repository,
	sessionsRepository *sessions.Repository,
	tokenManager *auth.TokenManager,
) *Service {
	return &Service{
		usersRepository:    usersRepository,
		sessionsRepository: sessionsRepository,
		tokenManager:       tokenManager,
	}
}

type ClientInfo struct {
	IPAddress *string
	UserAgent *string
}

type Result struct {
	User         *users.User `json:"user"`
	AccessToken  string      `json:"accessToken"`
	RefreshToken string      `json:"refreshToken"`
}

func ensureUserCanAuthenticate(user *users.User) error {
	if user.DeletedAt != nil {
		return ErrAccountDeleted
	}

	if user.InactivatedAt != nil {
		return ErrAccountInactive
	}

	return nil
}

func (service *Service) generateTokens(user *users.User) (string, string, error) {
	accessToken, err := service.tokenManager.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := service.tokenManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (service *Service) newRefreshTokenParams(userID string, refreshToken string, clientInfo ClientInfo) sessions.CreateParams {
	return sessions.CreateParams{
		UserID:    userID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(service.tokenManager.RefreshTokenTTL()),
		UserAgent: clientInfo.UserAgent,
		IPAddress: clientInfo.IPAddress,
	}
}

func (service *Service) Login(ctx context.Context, email string, password string, clientInfo ClientInfo) (*Result, error) {
	user, err := service.usersRepository.FindByEmail(ctx, users.NormalizeEmail(email))
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := ensureUserCanAuthenticate(user); err != nil {
		return nil, err
	}

	accessToken, refreshToken, err := service.generateTokens(user)
	if err != nil {
		return nil, err
	}

	err = service.sessionsRepository.Create(ctx, service.newRefreshTokenParams(user.ID, refreshToken, clientInfo))
	if err != nil {
		return nil, err
	}

	return &Result{User: user, AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (service *Service) Refresh(ctx context.Context, currentRefreshToken string, clientInfo ClientInfo) (*Result, error) {
	claims, err := service.tokenManager.ParseRefreshToken(currentRefreshToken)
	if err != nil {
		return nil, ErrRefreshTokenInvalidOrExpired
	}

	if claims.Subject == "" {
		return nil, ErrRefreshTokenMalformed
	}

	storedToken, err := service.sessionsRepository.FindByToken(ctx, currentRefreshToken)
	if err != nil {
		return nil, err
	}

	if storedToken == nil || storedToken.Revoked || storedToken.ExpiresAt.Before(time.Now()) || storedToken.UserID != claims.Subject {
		return nil, ErrRefreshTokenInvalid
	}

	user, err := service.usersRepository.FindByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	if err := ensureUserCanAuthenticate(user); err != nil {
		return nil, err
	}

	accessToken, refreshToken, err := service.generateTokens(user)
	if err != nil {
		return nil, err
	}

	err = service.sessionsRepository.Rotate(ctx, currentRefreshToken, service.newRefreshTokenParams(user.ID, refreshToken, clientInfo))
	if errors.Is(err, sessions.ErrRefreshTokenNotActive) {
		return nil, ErrRefreshTokenInvalid
	}

	if err != nil {
		return nil, err
	}

	return &Result{User: user, AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (service *Service) Logout(ctx context.Context, refreshToken string, userID string) error {
	if refreshToken == "" {
		return nil
	}

	return service.sessionsRepository.Revoke(ctx, refreshToken, userID)
}

func (service *Service) LogoutAll(ctx context.Context, userID string) error {
	return service.sessionsRepository.RevokeAllByUser(ctx, userID)
}
