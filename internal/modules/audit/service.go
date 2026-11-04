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

func (service *Service) Log(ctx context.Context, entry Entry) {
	logCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	if err := service.repository.Create(logCtx, entry); err != nil {
		log.Printf("Erro ao registrar log de auditoria (%s): %v", entry.Action, err)
	}
}
