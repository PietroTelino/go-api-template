package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port       int
	DBHost     string
	DBPort     int
	DBName     string
	DBUser     string
	DBPassword string

	JWTSecret                  string
	JWTExpiresInSeconds        int
	JWTRefreshSecret           string
	JWTRefreshExpiresInSeconds int
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Nenhum .env ecnontrado, usando variáveis do ambiente")
	}

	return &Config{
		Port:       getInt("PORT", 3333),
		DBHost:     require("DB_HOST"),
		DBPort:     getInt("DB_PORT", 5432),
		DBName:     require("DB_NAME"),
		DBUser:     require("DB_USER"),
		DBPassword: require("DB_PASSWORD"),

		JWTSecret:                  require("JWT_SECRET"),
		JWTExpiresInSeconds:        getInt("JWT_EXPIRES_IN_SECONDS", 900),
		JWTRefreshSecret:           require("JWT_REFRESH_SECRET"),
		JWTRefreshExpiresInSeconds: getInt("JWT_REFRESH_EXPIRES_IN_SECONDS", 604800),
	}
}

func require(Key string) string {
	variable := os.Getenv(Key)

	if variable == "" {
		log.Fatalf("Variável de ambiente obrigatória não definida: %s", Key)
	}

	return variable
}

func getInt(Key string, fallback int) int {
	variable := os.Getenv(Key)

	if variable == "" {
		return fallback
	}

	number, err := strconv.Atoi(variable)

	if err != nil {
		return fallback
	}

	return number
}
