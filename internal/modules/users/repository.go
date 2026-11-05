package users

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	userColumns         = `id, name, email, password, role, inactivated_at, deleted_at, created_at, updated_at, preferences`
	uniqueViolationCode = "23505"
)

type CreateParams struct {
	Name        string
	Email       string
	Password    string
	Role        string
	Preferences map[string]any
}

type Repository struct {
	dbPool *pgxpool.Pool
}

func NewRepository(dbPool *pgxpool.Pool) *Repository {
	return &Repository{dbPool: dbPool}
}

func scanUser(row pgx.Row) (*User, error) {
	user := &User{}

	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.InactivatedAt,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Preferences,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (repository *Repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT ` + userColumns + ` FROM users WHERE email = $1`

	user, err := scanUser(repository.dbPool.QueryRow(ctx, query, email))

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (repository *Repository) FindByID(ctx context.Context, id string) (*User, error) {
	query := `SELECT ` + userColumns + ` FROM users WHERE id = $1`

	user, err := scanUser(repository.dbPool.QueryRow(ctx, query, id))

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (repository *Repository) Create(ctx context.Context, params CreateParams) (*User, error) {
	query := `
		INSERT INTO users (name, email, password, role, preferences)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING ` + userColumns

	user, err := scanUser(repository.dbPool.QueryRow(
		ctx,
		query,
		params.Name,
		params.Email,
		params.Password,
		params.Role,
		params.Preferences,
	))

	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == uniqueViolationCode {
		return nil, ErrEmailAlreadyInUse
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}
