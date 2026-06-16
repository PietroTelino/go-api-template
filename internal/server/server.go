package server

import (
	"net/http"

	"github.com/PietroTelino/go-api-template/internal/modules/users"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(dbPool *pgxpool.Pool) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Get("/health", func(writer http.ResponseWriter, requet *http.Request) {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("ok"))
	})

	usersRepository := users.NewRepository(dbPool)
	usersService := users.NewService(usersRepository)
	usersHandler := users.NewHandler(usersService)

	router.Post("/users", usersHandler.Register)

	return router
}
