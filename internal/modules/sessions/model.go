package sessions

import (
	"errors"
	"time"
)

var ErrRefreshTokenNotActive = errors.New("auth.refreshTokenInvalid")

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

type CreateParams struct {
	UserID    string
	Token     string
	ExpiresAt time.Time
	UserAgent *string
	IPAddress *string
}
