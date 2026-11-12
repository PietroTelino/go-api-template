package audit

import (
	"context"
	"log"
	"time"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (service *Service) List(ctx context.Context, filters Filters) ([]*Record, int, error) {
	records, err := service.repository.FindAll(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	total, err := service.repository.Count(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

func (service *Service) GetByID(ctx context.Context, id string) (*Record, error) {
	return service.repository.FindByID(ctx, id)
}

func (service *Service) Log(ctx context.Context, entry Entry) {
	logCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	if err := service.repository.Create(logCtx, entry); err != nil {
		log.Printf("Erro ao registrar log de auditoria (%s): %v", entry.Action, err)
	}
}
