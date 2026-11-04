package server

import (
	"net/http"

	"github.com/PietroTelino/go-api-template/internal/auth"
	"github.com/PietroTelino/go-api-template/internal/config"
	"github.com/PietroTelino/go-api-template/internal/i18n"
	"github.com/PietroTelino/go-api-template/internal/modules/audit"
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

	tokenManager := auth.NewTokenManager(appConfig)
	_ = tokenManager

	router.Get("/health", func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.WriteHeader(http.StatusOK)
		responseWriter.Write([]byte("ok"))
	})

	auditRepository := audit.NewRepository(dbPool)
	auditService := audit.NewService(auditRepository)

	usersRepository := users.NewRepository(dbPool)
	usersService := users.NewService(usersRepository)
	usersHandler := users.NewHandler(usersService, auditService, translator)

	router.Route("/api", func(apiRouter chi.Router) {
		apiRouter.Post("/users", usersHandler.Register)
	})

	return router
}
