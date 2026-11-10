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

type UpdateParams struct {
	Name  *string
	Email *string
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

func isUniqueViolation(err error) bool {
	var pgError *pgconn.PgError
	return errors.As(err, &pgError) && pgError.Code == uniqueViolationCode
}

func (repository *Repository) findOne(ctx context.Context, query string, args ...any) (*User, error) {
	user, err := scanUser(repository.dbPool.QueryRow(ctx, query, args...))

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (repository *Repository) findMany(ctx context.Context, query string, args ...any) ([]*User, error) {
	rows, err := repository.dbPool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	foundUsers := []*User{}

	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}

		foundUsers = append(foundUsers, user)
	}

	return foundUsers, rows.Err()
}

func (repository *Repository) execByID(ctx context.Context, query string, id string, args ...any) error {
	_, err := repository.dbPool.Exec(ctx, query, append([]any{id}, args...)...)
	return err
}

func (repository *Repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	return repository.findOne(ctx, `SELECT `+userColumns+` FROM users WHERE email = $1`, email)
}

func (repository *Repository) FindByID(ctx context.Context, id string) (*User, error) {
	return repository.findOne(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id)
}

func (repository *Repository) FindAllActive(ctx context.Context) ([]*User, error) {
	return repository.findMany(ctx, `SELECT `+userColumns+` FROM users WHERE deleted_at IS NULL ORDER BY created_at ASC`)
}

func (repository *Repository) FindAllDeleted(ctx context.Context) ([]*User, error) {
	return repository.findMany(ctx, `SELECT `+userColumns+` FROM users WHERE deleted_at IS NOT NULL ORDER BY deleted_at DESC`)
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

	if isUniqueViolation(err) {
		return nil, ErrEmailAlreadyInUse
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (repository *Repository) Update(ctx context.Context, id string, params UpdateParams) (*User, error) {
	query := `
		UPDATE users
		SET name = COALESCE($2, name), email = COALESCE($3, email), updated_at = now()
		WHERE id = $1
		RETURNING ` + userColumns

	user, err := repository.findOne(ctx, query, id, params.Name, params.Email)

	if isUniqueViolation(err) {
		return nil, ErrEmailAlreadyInUse
	}

	return user, err
}

func (repository *Repository) UpdatePassword(ctx context.Context, id string, hashedPassword string) error {
	return repository.execByID(ctx, `UPDATE users SET password = $2, updated_at = now() WHERE id = $1`, id, hashedPassword)
}

func (repository *Repository) UpdateTheme(ctx context.Context, id string, theme string) error {
	query := `
		UPDATE users
		SET preferences = COALESCE(preferences, '{}'::jsonb) || jsonb_build_object('theme', $2::text), updated_at = now()
		WHERE id = $1
	`

	return repository.execByID(ctx, query, id, theme)
}

func (repository *Repository) SoftDelete(ctx context.Context, id string) error {
	return repository.execByID(ctx, `UPDATE users SET deleted_at = now(), updated_at = now() WHERE id = $1`, id)
}

func (repository *Repository) Restore(ctx context.Context, id string) error {
	return repository.execByID(ctx, `UPDATE users SET deleted_at = NULL, updated_at = now() WHERE id = $1`, id)
}

func (repository *Repository) Inactivate(ctx context.Context, id string) error {
	return repository.execByID(ctx, `UPDATE users SET inactivated_at = now(), updated_at = now() WHERE id = $1`, id)
}

func (repository *Repository) Reactivate(ctx context.Context, id string) error {
	return repository.execByID(ctx, `UPDATE users SET inactivated_at = NULL, updated_at = now() WHERE id = $1`, id)
}

func (repository *Repository) HardDelete(ctx context.Context, id string) error {
	return repository.execByID(ctx, `DELETE FROM users WHERE id = $1`, id)
}
