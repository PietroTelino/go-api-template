package passwordreset

import "errors"

var ErrInvalidOrExpiredToken = errors.New("password.invalidOrExpiredToken")
