package sessions

import (
	"errors"
	"time"
)

var (
	ErrRefreshTokenNotActive = errors.New("auth.refreshTokenInvalid")
	ErrSessionNotFound       = errors.New("user.sessionNotFound")
)

type RefreshToken struct {
	ID        string
	UserID    string
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
	Revoked   bool
	UserAgent *string
	IPAddress *string
}

type Session struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	UserAgent *string   `json:"userAgent"`
	IPAddress *string   `json:"ipAddress"`
}

type CreateParams struct {
	UserID    string
	Token     string
	ExpiresAt time.Time
	UserAgent *string
	IPAddress *string
}
