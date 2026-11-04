package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/PietroTelino/go-api-template/internal/config"
	"github.com/PietroTelino/go-api-template/internal/database"
	"github.com/PietroTelino/go-api-template/migrations"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Falha ao executar as migrations: %v", err)
	}
}

func run() error {
	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	appConfig := config.Load()

	dbPool, err := database.Connect(appConfig)
	if err != nil {
		return err
	}
	defer dbPool.Close()

	sqlDB := stdlib.OpenDBFromPool(dbPool)
	defer sqlDB.Close()

	provider, err := goose.NewProvider(goose.DialectPostgres, sqlDB, migrations.Files)
	if err != nil {
		return fmt.Errorf("erro ao criar o provider do goose: %w", err)
	}

	ctx := context.Background()

	switch command {
	case "up":
		return runUp(ctx, provider)
	case "down":
		return runDown(ctx, provider)
	case "status":
		return runStatus(ctx, provider)
	default:
		return fmt.Errorf("comando desconhecido %q (use up, down ou status)", command)
	}
}

func runUp(ctx context.Context, provider *goose.Provider) error {
	results, err := provider.Up(ctx)
	if err != nil {
		return fmt.Errorf("erro ao aplicar as migrations: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("Nenhuma migration pendente.")
		return nil
	}

	for _, result := range results {
		fmt.Println(result)
	}

	return nil
}

func runDown(ctx context.Context, provider *goose.Provider) error {
	result, err := provider.Down(ctx)
	if err != nil {
		return fmt.Errorf("erro ao reverter a migration: %w", err)
	}

	fmt.Println(result)

	return nil
}

func runStatus(ctx context.Context, provider *goose.Provider) error {
	statuses, err := provider.Status(ctx)
	if err != nil {
		return fmt.Errorf("erro ao consultar o status das migrations: %w", err)
	}

	for _, status := range statuses {
		fmt.Printf("%-8s %s\n", status.State, status.Source.Path)
	}

	return nil
}
