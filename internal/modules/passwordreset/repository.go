package passwordreset

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	dbPool *pgxpool.Pool
}

func NewRepository(dbPool *pgxpool.Pool) *Repository {
	return &Repository{dbPool: dbPool}
}

func (repository *Repository) Create(ctx context.Context, userID string, tokenHash string, expiresAt time.Time) error {
	return pgx.BeginFunc(ctx, repository.dbPool, func(transaction pgx.Tx) error {
		_, err := transaction.Exec(
			ctx,
			`UPDATE password_resets SET used_at = now() WHERE user_id = $1 AND used_at IS NULL`,
			userID,
		)
		if err != nil {
			return err
		}

		_, err = transaction.Exec(
			ctx,
			`INSERT INTO password_resets (user_id, token, expires_at) VALUES ($1, $2, $3)`,
			userID,
			tokenHash,
			expiresAt,
		)

		return err
	})
}

func (repository *Repository) Consume(ctx context.Context, tokenHash string, hashedPassword string) (userID string, email string, err error) {
	err = pgx.BeginFunc(ctx, repository.dbPool, func(transaction pgx.Tx) error {
		err := transaction.QueryRow(
			ctx,
			`UPDATE password_resets
			SET used_at = now()
			WHERE token = $1 AND used_at IS NULL AND expires_at > now()
			RETURNING user_id`,
			tokenHash,
		).Scan(&userID)

		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidOrExpiredToken
		}

		if err != nil {
			return err
		}

		err = transaction.QueryRow(
			ctx,
			`UPDATE users
			SET password = $2, updated_at = now()
			WHERE id = $1 AND deleted_at IS NULL
			RETURNING email`,
			userID,
			hashedPassword,
		).Scan(&email)

		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidOrExpiredToken
		}

		return err
	})

	return userID, email, err
}
