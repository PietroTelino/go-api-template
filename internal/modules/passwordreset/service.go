package passwordreset

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/PietroTelino/go-api-template/internal/modules/users"
	"golang.org/x/crypto/bcrypt"
)

type SessionRevoker interface {
	RevokeAllByUser(ctx context.Context, userID string) error
}

type ResetEmailSender interface {
	SendPasswordReset(to string, resetLink string) error
}

type Service struct {
	usersRepository *users.Repository
	repository      *Repository
	sessionRevoker  SessionRevoker
	emailSender     ResetEmailSender
	tokenTTL        time.Duration
	frontendURL     string
}

func NewService(
	usersRepository *users.Repository,
	repository *Repository,
	sessionRevoker SessionRevoker,
	emailSender ResetEmailSender,
	tokenTTL time.Duration,
	frontendURL string,
) *Service {
	return &Service{
		usersRepository: usersRepository,
		repository:      repository,
		sessionRevoker:  sessionRevoker,
		emailSender:     emailSender,
		tokenTTL:        tokenTTL,
		frontendURL:     frontendURL,
	}
}

type ConfirmResult struct {
	UserID string
	Email  string
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (service *Service) buildResetLink(token string) string {
	query := url.Values{"token": {token}}
	return strings.TrimRight(service.frontendURL, "/") + "/reset-password?" + query.Encode()
}

func (service *Service) RequestReset(ctx context.Context, email string) error {
	user, err := service.usersRepository.FindByEmail(ctx, users.NormalizeEmail(email))
	if err != nil {
		return err
	}

	if user == nil || user.DeletedAt != nil {
		return nil
	}

	token := rand.Text()

	if err := service.repository.Create(ctx, user.ID, hashToken(token), time.Now().Add(service.tokenTTL)); err != nil {
		return err
	}

	if err := service.emailSender.SendPasswordReset(user.Email, service.buildResetLink(token)); err != nil {
		log.Printf("Erro ao enviar o e-mail de redefinição de senha para %s: %v", user.Email, err)
	}

	return nil
}

func (service *Service) ConfirmReset(ctx context.Context, token string, newPassword string) (*ConfirmResult, error) {
	if err := users.ValidatePassword(newPassword); err != nil {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userID, email, err := service.repository.Consume(ctx, hashToken(token), string(hashedPassword))
	if err != nil {
		return nil, err
	}

	if err := service.sessionRevoker.RevokeAllByUser(ctx, userID); err != nil {
		return nil, err
	}

	return &ConfirmResult{UserID: userID, Email: email}, nil
}
