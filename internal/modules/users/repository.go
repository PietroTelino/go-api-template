package users

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	dbPool *pgxpool.Pool
}

func NewRepository(dbPool *pgxpool.Pool) *Repository {
	return &Repository{dbPool: dbPool}
}

func (repository *Repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, name, email, password, role, inactivated_at, deleted_at, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &User{}

	err := repository.dbPool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.InactivatedAt,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (repository *Repository) Create(ctx context.Context, name string, email string, hashedPassword string) (*User, error) {
	query := `
		INSERT INTO users (name, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, password, role, inactivated_at, deleted_at, created_at, updated_at
	`

	user := &User{}

	err := repository.dbPool.QueryRow(ctx, query, name, email, hashedPassword).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.InactivatedAt,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}
