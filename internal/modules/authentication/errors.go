package authentication

import "errors"

var (
	ErrInvalidCredentials           = errors.New("auth.invalidCredentials")
	ErrAccountDeleted               = errors.New("auth.accountDeleted")
	ErrAccountInactive              = errors.New("auth.accountInactive")
	ErrRefreshTokenInvalidOrExpired = errors.New("auth.refreshTokenInvalidOrExpired")
	ErrRefreshTokenMalformed        = errors.New("auth.refreshTokenMalformed")
	ErrRefreshTokenInvalid          = errors.New("auth.refreshTokenInvalid")
	ErrUserNotFound                 = errors.New("user.notFound")
)
