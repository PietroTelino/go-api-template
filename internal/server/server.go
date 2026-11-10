package server

import (
	"net/http"

	"github.com/PietroTelino/go-api-template/internal/auth"
	"github.com/PietroTelino/go-api-template/internal/config"
	"github.com/PietroTelino/go-api-template/internal/i18n"
	"github.com/PietroTelino/go-api-template/internal/modules/audit"
	"github.com/PietroTelino/go-api-template/internal/modules/authentication"
	"github.com/PietroTelino/go-api-template/internal/modules/sessions"
	"github.com/PietroTelino/go-api-template/internal/modules/users"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(dbPool *pgxpool.Pool, appConfig *config.Config, translator *i18n.Translator) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{appConfig.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept-Language", "Authorization", "Content-Type", "X-Refresh-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	router.Use(i18n.Middleware(translator))

	router.Get("/health", func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.WriteHeader(http.StatusOK)
		responseWriter.Write([]byte("ok"))
	})

	tokenManager := auth.NewTokenManager(appConfig)
	requireAuthentication := auth.RequireAuthentication(tokenManager, translator)
	requireAdmin := auth.RequiredRoles(translator, users.RoleAdministrator, users.RoleGod)
	requireGod := auth.RequiredRoles(translator, users.RoleGod)

	auditRepository := audit.NewRepository(dbPool)
	auditService := audit.NewService(auditRepository)

	sessionsRepository := sessions.NewRepository(dbPool)

	usersRepository := users.NewRepository(dbPool)
	usersService := users.NewService(usersRepository, sessionsRepository)
	usersHandler := users.NewHandler(usersService, auditService, translator)

	authenticationService := authentication.NewService(usersRepository, sessionsRepository, tokenManager)
	authenticationHandler := authentication.NewHandler(authenticationService, auditService, translator)

	router.Route("/api", func(apiRouter chi.Router) {
		apiRouter.Route("/users", func(usersRouter chi.Router) {
			usersRouter.Post("/", usersHandler.Register)

			usersRouter.Group(func(authenticatedRouter chi.Router) {
				authenticatedRouter.Use(requireAuthentication)

				authenticatedRouter.Patch("/me/preferences", usersHandler.UpdateMyPreferences)
				authenticatedRouter.Patch("/me/password", usersHandler.ChangeMyPassword)
				authenticatedRouter.Delete("/me/delete", usersHandler.SelfDelete)

				authenticatedRouter.Group(func(adminRouter chi.Router) {
					adminRouter.Use(requireAdmin)

					adminRouter.Get("/", usersHandler.List)
					adminRouter.Get("/deleted-users", usersHandler.ListDeleted)
					adminRouter.Get("/{id}", usersHandler.GetByID)
					adminRouter.Post("/create", usersHandler.Create)
					adminRouter.Put("/{id}", usersHandler.Update)
					adminRouter.Delete("/{id}", usersHandler.Delete)
					adminRouter.Patch("/{id}/restore", usersHandler.Restore)
					adminRouter.Patch("/{id}/inactivate", usersHandler.Inactivate)
					adminRouter.Patch("/{id}/reactivate", usersHandler.Reactivate)

					adminRouter.With(requireGod).Delete("/{id}/hard", usersHandler.HardDelete)
				})
			})
		})

		apiRouter.Route("/auth", func(authRouter chi.Router) {
			authRouter.Post("/login", authenticationHandler.Login)
			authRouter.Post("/refresh", authenticationHandler.Refresh)

			authRouter.With(requireAuthentication).Post("/logout", authenticationHandler.Logout)
			authRouter.With(requireAuthentication).Post("/logout-all", authenticationHandler.LogoutAll)
		})
	})

	return router
}
