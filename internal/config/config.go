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

	EmailEnabled  bool
	EmailHost     string
	EmailPort     int
	EmailSecure   bool
	EmailUser     string
	EmailPassword string
	EmailFrom     string

	FrontendURL string

	PasswordResetExpiresMS int
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Nenhum .env encontrado, usando variáveis do ambiente")
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

		EmailEnabled:  getBool("EMAIL_ENABLED", false),
		EmailHost:     os.Getenv("EMAIL_HOST"),
		EmailPort:     getInt("EMAIL_PORT", 587),
		EmailSecure:   getBool("EMAIL_SECURE", false),
		EmailUser:     os.Getenv("EMAIL_USER"),
		EmailPassword: os.Getenv("EMAIL_PASSWORD"),
		EmailFrom:     getString("EMAIL_FROM", "no-reply@localhost"),

		FrontendURL: require("FRONTEND_URL"),

		PasswordResetExpiresMS: getInt("PASSWORD_RESET_EXPIRES_MS", 3600000),
	}
}

func require(key string) string {
	variable := os.Getenv(key)

	if variable == "" {
		log.Fatalf("Variável de ambiente obrigatória não definida: %s", key)
	}

	return variable
}

func getString(key string, fallback string) string {
	variable := os.Getenv(key)

	if variable == "" {
		return fallback
	}

	return variable
}

func getInt(key string, fallback int) int {
	variable := os.Getenv(key)

	if variable == "" {
		return fallback
	}

	number, err := strconv.Atoi(variable)

	if err != nil {
		return fallback
	}

	return number
}

func getBool(key string, fallback bool) bool {
	variable := os.Getenv(key)

	if variable == "" {
		return fallback
	}

	boolean, err := strconv.ParseBool(variable)

	if err != nil {
		return fallback
	}

	return boolean
}
