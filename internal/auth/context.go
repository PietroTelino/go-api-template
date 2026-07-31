package auth

import "context"

type contextKey string

const authenticatedUserContextKey contextKey = "authenticatedUser"

type AuthenticatedUser struct {
	ID          string
	Role        string
	AccessToken string
}

func WithAuthenticatedUser(ctx context.Context, authenticatedUser *AuthenticatedUser) context.Context {
	return context.WithValue(ctx, authenticatedUserContextKey, authenticatedUser)
}

func AuthenticatedUserFromContext(ctx context.Context) (*AuthenticatedUser, bool) {
	authenticatedUser, isAuthenticated := ctx.Value(authenticatedUserContextKey).(*AuthenticatedUser)
	return authenticatedUser, isAuthenticated
}
