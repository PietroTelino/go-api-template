package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/PietroTelino/go-api-template/internal/config"
	"github.com/PietroTelino/go-api-template/internal/database"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Falha ao executar o seed: %v", err)
	}
}

func run() error {
	appConfig := config.Load()

	godEmail := os.Getenv("GOD_EMAIL")
	godPassword := os.Getenv("GOD_PASSWORD")

	if !strings.Contains(godEmail, "@") {
		return fmt.Errorf("GOD_EMAIL inválido")
	}

	if godPassword == "" {
		return fmt.Errorf("GOD_PASSWORD não definida")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(godPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("erro ao gerar o hash da senha: %w", err)
	}

	dbPool, err := database.Connect(appConfig)
	if err != nil {
		return err
	}
	defer dbPool.Close()

	query := `
		INSERT INTO users (name, email, password, role)
		VALUES ('God User', $1, $2, 'god')
		ON CONFLICT (email) DO NOTHING
	`

	commandTag, err := dbPool.Exec(context.Background(), query, godEmail, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("erro ao criar o usuário GOD: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		fmt.Printf("Usuário GOD já existia (%s). Nada foi alterado.\n", godEmail)
		return nil
	}

	fmt.Println("Usuário GOD criado com sucesso!")

	return nil
}
