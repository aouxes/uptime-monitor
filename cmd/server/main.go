package main

import (
	"context"
	"log"
	"net/http"

	"github.com/aouxes/uptime-monitor/internal/config"
	"github.com/aouxes/uptime-monitor/internal/models"
	"github.com/aouxes/uptime-monitor/internal/storage"
)

func main() {
	cfg := config.Load()
	db, err := storage.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// ТЕСТ: Создаем test user
	ctx := context.Background()
	testUser := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password_123", // Пока просто заглушка
	}

	if err := db.CreateUser(ctx, testUser); err != nil {
		log.Printf("Warning: Failed to create test user: %v", err)
	} else {
		log.Printf("Test user created: ID=%d", testUser.ID)
	}

	// Запускаем сервер
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<h1>Uptime Monitor is running!</h1>"))
		w.Write([]byte("<p>Database connection: ✅ OK</p>"))
		w.Write([]byte("<p>Try to create a user via API soon!</p>"))
	})

	log.Printf("Server starting on :%s...", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, nil); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}
}
