package sessions

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const insertRefreshTokenQuery = `
	INSERT INTO refresh_tokens (user_id, token, expires_at, user_agent, ip_address)
	VALUES ($1, $2, $3, $4, $5)
`

type Repository struct {
	dbPool *pgxpool.Pool
}

func NewRepository(dbPool *pgxpool.Pool) *Repository {
	return &Repository{dbPool: dbPool}
}

func (repository *Repository) Create(ctx context.Context, params CreateParams) error {
	_, err := repository.dbPool.Exec(
		ctx,
		insertRefreshTokenQuery,
		params.UserID,
		params.Token,
		params.ExpiresAt,
		params.UserAgent,
		params.IPAddress,
	)

	return err
}

func (repository *Repository) FindByToken(ctx context.Context, token string) (*RefreshToken, error) {
	query := `
		SELECT id, user_id, token, created_at, expires_at, revoked, user_agent, ip_address
		FROM refresh_tokens
		WHERE token = $1
	`

	refreshToken := &RefreshToken{}

	err := repository.dbPool.QueryRow(ctx, query, token).Scan(
		&refreshToken.ID,
		&refreshToken.UserID,
		&refreshToken.Token,
		&refreshToken.CreatedAt,
		&refreshToken.ExpiresAt,
		&refreshToken.Revoked,
		&refreshToken.UserAgent,
		&refreshToken.IPAddress,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return refreshToken, nil
}

func (repository *Repository) FindActiveByUser(ctx context.Context, userID string) ([]*Session, error) {
	query := `
		SELECT id, created_at, expires_at, user_agent, ip_address
		FROM refresh_tokens
		WHERE user_id = $1 AND revoked = false AND expires_at > now()
		ORDER BY created_at DESC
	`

	rows, err := repository.dbPool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	activeSessions := []*Session{}

	for rows.Next() {
		session := &Session{}

		if err := rows.Scan(&session.ID, &session.CreatedAt, &session.ExpiresAt, &session.UserAgent, &session.IPAddress); err != nil {
			return nil, err
		}

		activeSessions = append(activeSessions, session)
	}

	return activeSessions, rows.Err()
}

func (repository *Repository) RevokeByIDForUser(ctx context.Context, sessionID string, userID string) (bool, error) {
	commandTag, err := repository.dbPool.Exec(
		ctx,
		`UPDATE refresh_tokens
		SET revoked = true
		WHERE id = $1 AND user_id = $2 AND revoked = false AND expires_at > now()`,
		sessionID,
		userID,
	)

	if err != nil {
		return false, err
	}

	return commandTag.RowsAffected() > 0, nil
}

func (repository *Repository) Rotate(ctx context.Context, currentToken string, newToken CreateParams) error {
	return pgx.BeginFunc(ctx, repository.dbPool, func(transaction pgx.Tx) error {
		commandTag, err := transaction.Exec(
			ctx,
			`UPDATE refresh_tokens SET revoked = true WHERE token = $1 AND revoked = false`,
			currentToken,
		)
		if err != nil {
			return err
		}

		if commandTag.RowsAffected() == 0 {
			return ErrRefreshTokenNotActive
		}

		_, err = transaction.Exec(
			ctx,
			insertRefreshTokenQuery,
			newToken.UserID,
			newToken.Token,
			newToken.ExpiresAt,
			newToken.UserAgent,
			newToken.IPAddress,
		)

		return err
	})
}

func (repository *Repository) Revoke(ctx context.Context, token string, userID string) error {
	_, err := repository.dbPool.Exec(
		ctx,
		`UPDATE refresh_tokens SET revoked = true WHERE token = $1 AND user_id = $2 AND revoked = false`,
		token,
		userID,
	)

	return err
}

func (repository *Repository) RevokeAllByUser(ctx context.Context, userID string) error {
	_, err := repository.dbPool.Exec(
		ctx,
		`UPDATE refresh_tokens SET revoked = true WHERE user_id = $1 AND revoked = false`,
		userID,
	)

	return err
}
