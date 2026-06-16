package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PietroTelino/go-api-template/internal/config"
	"github.com/PietroTelino/go-api-template/internal/database"
	"github.com/PietroTelino/go-api-template/internal/server"
)

func main() {
	appConfig := config.Load()

	dbPool, err := database.Connect(appConfig)
	if err != nil {
		log.Fatalf("Falha ao conectar ao banco: %v", err)
	}

	defer dbPool.Close()

	log.Println("Banco de dados conectado com sucesso!")

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", appConfig.Port),
		Handler: server.New(dbPool),
	}

	go func() {
		log.Printf("Servidor rodando em http://localhost%s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro ao iniciar servidor: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Encerrando servidor...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Erro ao encerrar servidor: %v", err)
	}

	log.Println("Servidor encerrado.")
}
