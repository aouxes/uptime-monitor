package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     int
	DBName     string
	DBUser     string
	DBPassword string
	ServerPort string
	JWTSecret  string
}

func Load() *Config {
	err := godotenv.Load("deployments/.env")
	if err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		log.Fatalf("Invalid DB_PORT: %v", err)
	}

	dbPassword := getEnv("DB_PASSWORD", "")
	if dbPassword == "" {
		log.Fatal("DB_PASSWORD is required and cannot be empty")
	}

	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     dbPort,
		DBName:     getEnv("DB_NAME", "uptimedb"),
		DBUser:     getEnv("DB_USER", "monitoruser"),
		DBPassword: dbPassword,
		ServerPort: getEnv("SERVER_PORT", "8080"),
		JWTSecret:  getEnv("JWT_SECRET", "your-super-secret-key-here"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
