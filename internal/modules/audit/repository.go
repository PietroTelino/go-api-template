package audit

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	dbPool *pgxpool.Pool
}

func NewRepository(dbPool *pgxpool.Pool) *Repository {
	return &Repository{dbPool: dbPool}
}

func (repository *Repository) Create(ctx context.Context, entry Entry) error {
	query := `
		INSERT INTO audit_logs (user_id, performed_by_id, action, entity, entity_id, details, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := repository.dbPool.Exec(
		ctx,
		query,
		entry.UserID,
		entry.PerformedByID,
		string(entry.Action),
		string(entry.Entity),
		entry.EntityID,
		entry.Details,
		entry.IPAddress,
		entry.UserAgent,
	)

	return err
}
