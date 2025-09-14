package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost        string
	DBPort        int
	DBName        string
	DBUser        string
	DBPassword    string
	ServerPort    string
	JWTSecret     string
	TelegramToken string
}

func Load() *Config {
	// Попробуем загрузить .env файл из разных мест
	envFiles := []string{"deployments/.env", ".env", "../deployments/.env"}
	for _, envFile := range envFiles {
		if err := godotenv.Load(envFile); err == nil {
			log.Printf("Loaded .env file from: %s", envFile)
			break
		}
	}

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		log.Fatalf("Invalid DB_PORT: %v", err)
	}

	dbPassword := getEnv("DB_PASSWORD", "defaultpassword")
	if dbPassword == "" {
		log.Fatal("DB_PASSWORD is required and cannot be empty")
	}

	jwtSecret := getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production")
	if jwtSecret == "your-super-secret-key-here" {
		log.Printf("Warning: Using default JWT secret. Please set JWT_SECRET in environment variables for production!")
	}

	return &Config{
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        dbPort,
		DBName:        getEnv("DB_NAME", "uptimedb"),
		DBUser:        getEnv("DB_USER", "monitoruser"),
		DBPassword:    dbPassword,
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		JWTSecret:     jwtSecret,
		TelegramToken: getEnv("TELEGRAM_TOKEN", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
