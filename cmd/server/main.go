package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aouxes/uptime-monitor/internal/checker"
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

	checker := checker.New(db, 5*time.Minute)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go checker.Start(ctx)

	userHandler := handlers.NewUserHandler(db)
	siteHandler := handlers.NewSiteHandler(db)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/register", userHandler.Register)
	mux.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
		userHandler.Login(w, r, cfg.JWTSecret)
	})
	mux.Handle("POST /api/sites", middleware.AuthMiddleware(cfg.JWTSecret)(http.HandlerFunc(siteHandler.AddSite)))
	mux.Handle("GET /api/sites", middleware.AuthMiddleware(cfg.JWTSecret)(http.HandlerFunc(siteHandler.GetSites)))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<h1>Uptime Monitor API</h1>"))
		w.Write([]byte("<p>Endpoints: POST /api/register, POST /api/login, POST /api/sites, GET /api/sites</p>"))
		w.Write([]byte("<p>Background checker is running every 5 minutes</p>"))
	})

	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: mux,
	}

	go func() {
		log.Printf("Server starting on :%s...", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server startup failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Printf("Shutting down server...")

	cancel()
	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Printf("Server stopped")
}
