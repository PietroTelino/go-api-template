package users

import (
	"time"
)

const (
	RoleUser          = "user"
	RoleAdministrator = "administrator"
	RoleGod           = "god"
)

type User struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Email         string         `json:"email"`
	Password      string         `json:"-"`
	Role          string         `json:"role"`
	InactivatedAt *time.Time     `json:"inactivatedAt,omitempty"`
	DeletedAt     *time.Time     `json:"deletedAt,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	Preferences   map[string]any `json:"preferences,omitempty"`
}
