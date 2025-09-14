package main

import (
	"log"
	"net/http"

	"github.com/aouxes/uptime-monitor/internal/config"
	"github.com/aouxes/uptime-monitor/internal/handlers"
	"github.com/aouxes/uptime-monitor/internal/middleware"
	"github.com/aouxes/uptime-monitor/internal/storage"
)

func main() {
	cfg := config.Load()
	db, err := storage.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	userHandler := handlers.NewUserHandler(db)
	siteHandler := handlers.NewSiteHandler(db)

	http.HandleFunc("POST /api/register", userHandler.Register)
	http.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
		userHandler.Login(w, r, cfg.JWTSecret)
	})

	protected := http.NewServeMux()
	protected.HandleFunc("POST /api/sites", siteHandler.AddSite)

	http.Handle("/api/", middleware.AuthMiddleware(cfg.JWTSecret)(protected))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<h1>Uptime Monitor API</h1>"))
		w.Write([]byte("<p>Endpoints: POST /api/register, POST /api/login, POST /api/sites</p>"))
	})

	log.Printf("Server starting on :%s...", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, nil); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}
}
