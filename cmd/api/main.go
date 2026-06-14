package main

import (
	"log"

	"github.com/PietroTelino/go-api-template/internal/config"
	"github.com/PietroTelino/go-api-template/internal/database"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Falha ao conectar ao banco: %v", err)
	}

	defer db.Close()

	log.Println("Banco de dados conectado com sucesso!")
}
