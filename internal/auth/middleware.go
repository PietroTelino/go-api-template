package auth

import (
	"net/http"
	"strings"

	"github.com/PietroTelino/go-api-template/internal/i18n"
	"github.com/PietroTelino/go-api-template/internal/response"
)

func RequireAuthentication(tokenManager *TokenManager, translator *i18n.Translator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			authorizationHeader := request.Header.Get("Authorization")
			if authorizationHeader == "" {
				response.WriteTranslatedError(responseWriter, request, translator, http.StatusUnauthorized, "auth.tokenNotProvided")
				return
			}

			authorizationParts := strings.SplitN(authorizationHeader, " ", 2)
			if len(authorizationParts) != 2 || authorizationParts[0] != "Bearer" || authorizationParts[1] == "" {
				response.WriteTranslatedError(responseWriter, request, translator, http.StatusUnauthorized, "auth.invalidTokenFormat")
				return
			}
			accessTokenString := authorizationParts[1]

			claims, err := tokenManager.ParseAccessToken(accessTokenString)
			if err != nil {
				response.WriteTranslatedError(responseWriter, request, translator, http.StatusUnauthorized, "auth.invalidOrExpiredToken")
				return
			}

			if claims.Subject == "" {
				response.WriteTranslatedError(responseWriter, request, translator, http.StatusUnauthorized, "auth.malformedToken")
				return
			}

			userRole := claims.Role
			if userRole == "" {
				userRole = "user"
			}

			authenticatedUser := &AuthenticatedUser{
				ID:          claims.Subject,
				Role:        userRole,
				AccessToken: accessTokenString,
			}

			requestWithUser := request.WithContext(WithAuthenticatedUser(request.Context(), authenticatedUser))
			next.ServeHTTP(responseWriter, requestWithUser)
		})
	}
}

func RequiredRoles(translator *i18n.Translator, allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			authenticatedUser, isAuthenticated := AuthenticatedUserFromContext(request.Context())
			if !isAuthenticated {
				response.WriteTranslatedError(responseWriter, request, translator, http.StatusUnauthorized, "auth.notAuthenticated")
				return
			}

			for _, allowedRole := range allowedRoles {
				if authenticatedUser.Role == allowedRole {
					next.ServeHTTP(responseWriter, request)
					return
				}
			}

			response.WriteTranslatedError(responseWriter, request, translator, http.StatusForbidden, "auth.accessDenied")
		})
	}
}
