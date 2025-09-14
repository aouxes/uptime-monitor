package main

import (
	"log"
	"net/http"

	"github.com/aouxes/uptime-monitor/internal/config"
	"github.com/aouxes/uptime-monitor/internal/handlers"
	"github.com/aouxes/uptime-monitor/internal/storage"
)

func main() {
	cfg := config.Load()
	db, err := storage.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Создаем обработчики
	userHandler := handlers.NewUserHandler(db)

	// Настраиваем роутинг
	http.HandleFunc("POST /api/register", userHandler.Register)

	// Старый обработчик для проверки работы
	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<h1>Uptime Monitor API</h1>"))
		w.Write([]byte("<p>Use POST /api/register to create user</p>"))
	})

	log.Printf("Server starting on :%s...", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, nil); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}
}
