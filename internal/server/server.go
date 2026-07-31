package server

import (
	"net/http"

	"github.com/PietroTelino/go-api-template/internal/auth"
	"github.com/PietroTelino/go-api-template/internal/config"
	"github.com/PietroTelino/go-api-template/internal/i18n"
	"github.com/PietroTelino/go-api-template/internal/modules/users"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(dbPool *pgxpool.Pool, appConfig *config.Config, translator *i18n.Translator) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(i18n.Middleware(translator))

	tokenManager := auth.NewTokenManager(appConfig)
	_ = tokenManager

	router.Get("/health", func(responseWriter http.ResponseWriter, requet *http.Request) {
		responseWriter.WriteHeader(http.StatusOK)
		responseWriter.Write([]byte("ok"))
	})

	usersRepository := users.NewRepository(dbPool)
	usersService := users.NewService(usersRepository)
	usersHandler := users.NewHandler(usersService, translator)

	router.Post("/users", usersHandler.Register)

	return router
}
