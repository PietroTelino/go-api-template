package sessions

import (
	"context"

	"github.com/PietroTelino/go-api-template/internal/validation"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (service *Service) ListMySessions(ctx context.Context, userID string) ([]*Session, error) {
	return service.repository.FindActiveByUser(ctx, userID)
}

func (service *Service) TerminateSession(ctx context.Context, userID string, sessionID string) error {
	if !validation.IsUUID(sessionID) {
		return ErrSessionNotFound
	}

	wasRevoked, err := service.repository.RevokeByIDForUser(ctx, sessionID, userID)
	if err != nil {
		return err
	}

	if !wasRevoked {
		return ErrSessionNotFound
	}

	return nil
}
