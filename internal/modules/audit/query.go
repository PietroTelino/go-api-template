package audit

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type UserSummary struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type Record struct {
	ID            string         `json:"id"`
	UserID        *string        `json:"userId"`
	PerformedByID *string        `json:"performedById"`
	Action        string         `json:"action"`
	Entity        string         `json:"entity"`
	EntityID      *string        `json:"entityId"`
	Details       map[string]any `json:"details"`
	IPAddress     *string        `json:"ipAddress"`
	UserAgent     *string        `json:"userAgent"`
	CreatedAt     time.Time      `json:"createdAt"`
	User          *UserSummary   `json:"user"`
	PerformedBy   *UserSummary   `json:"performedBy"`
}

type Filters struct {
	UserID        string
	PerformedByID string
	Action        string
	Entity        string
	Limit         int
	Offset        int
}

const recordSelect = `
	SELECT
		a.id, a.user_id, a.performed_by_id, a.action, a.entity, a.entity_id,
		a.details, a.ip_address, a.user_agent, a.created_at,
		u.id, u.name, u.email, u.role,
		p.id, p.name, p.email, p.role
	FROM audit_logs a
	LEFT JOIN users u ON u.id = a.user_id
	LEFT JOIN users p ON p.id = a.performed_by_id
`

type nullableUser struct {
	id    *string
	name  *string
	email *string
	role  *string
}

func (user nullableUser) summary() *UserSummary {
	if user.id == nil {
		return nil
	}

	return &UserSummary{ID: *user.id, Name: *user.name, Email: *user.email, Role: *user.role}
}

func scanRecord(row pgx.Row) (*Record, error) {
	record := &Record{}
	var targetUser, performer nullableUser

	err := row.Scan(
		&record.ID,
		&record.UserID,
		&record.PerformedByID,
		&record.Action,
		&record.Entity,
		&record.EntityID,
		&record.Details,
		&record.IPAddress,
		&record.UserAgent,
		&record.CreatedAt,
		&targetUser.id,
		&targetUser.name,
		&targetUser.email,
		&targetUser.role,
		&performer.id,
		&performer.name,
		&performer.email,
		&performer.role,
	)

	if err != nil {
		return nil, err
	}

	record.User = targetUser.summary()
	record.PerformedBy = performer.summary()

	return record, nil
}

func buildWhereClause(filters Filters) (string, []any) {
	conditions := []string{}
	args := []any{}

	addCondition := func(column string, value string) {
		if value == "" {
			return
		}

		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf("%s = $%d", column, len(args)))
	}

	addCondition("a.user_id", filters.UserID)
	addCondition("a.performed_by_id", filters.PerformedByID)
	addCondition("a.action", filters.Action)
	addCondition("a.entity", filters.Entity)

	if len(conditions) == 0 {
		return "", args
	}

	return " WHERE " + strings.Join(conditions, " AND "), args
}

func (repository *Repository) FindAll(ctx context.Context, filters Filters) ([]*Record, error) {
	whereClause, args := buildWhereClause(filters)

	args = append(args, filters.Limit, filters.Offset)
	query := recordSelect + whereClause + fmt.Sprintf(" ORDER BY a.created_at DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))

	rows, err := repository.dbPool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := []*Record{}

	for rows.Next() {
		record, err := scanRecord(rows)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	return records, rows.Err()
}

func (repository *Repository) Count(ctx context.Context, filters Filters) (int, error) {
	whereClause, args := buildWhereClause(filters)

	var total int
	err := repository.dbPool.QueryRow(ctx, `SELECT count(*) FROM audit_logs a`+whereClause, args...).Scan(&total)

	return total, err
}

func (repository *Repository) FindByID(ctx context.Context, id string) (*Record, error) {
	record, err := scanRecord(repository.dbPool.QueryRow(ctx, recordSelect+` WHERE a.id = $1`, id))

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return record, err
}
