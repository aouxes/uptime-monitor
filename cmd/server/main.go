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
	"github.com/aouxes/uptime-monitor/internal/notifier"
	"github.com/aouxes/uptime-monitor/internal/storage"
	"github.com/aouxes/uptime-monitor/internal/telegram"
)

func main() {
	cfg := config.Load()
	db, err := storage.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Создаем и запускаем checker с 20 workers
	checker := checker.New(db, 5*time.Minute, 20, cfg.TelegramToken)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Printf("Starting background site checker with 20 workers...")
	if cfg.TelegramToken != "" {
		log.Printf("Telegram notifications enabled")

		// Запускаем Telegram бот
		bot := telegram.NewBot(cfg.TelegramToken, db)
		go func() {
			if err := bot.Start(ctx); err != nil {
				log.Printf("Telegram bot error: %v", err)
			}
		}()
		log.Printf("Telegram bot started")
	} else {
		log.Printf("Telegram notifications disabled (no token provided)")
	}
	go checker.Start(ctx)

	// Создаем notifier для уведомлений
	notifier := notifier.New(cfg.TelegramToken, db)

	// Создаем обработчики
	userHandler := handlers.NewUserHandler(db)
	siteHandler := handlers.NewSiteHandler(db, notifier)

	mux := http.NewServeMux()

	// Обслуживаем статические файлы (CSS, JS)
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Обслуживаем HTML шаблон
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/index.html")
	})

	// API endpoints
	mux.HandleFunc("POST /api/register", userHandler.Register)
	mux.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
		userHandler.Login(w, r, cfg.JWTSecret)
	})

	// Защищенные endpoints (требуют JWT)
	mux.Handle("POST /api/sites", middleware.AuthMiddleware(cfg.JWTSecret)(http.HandlerFunc(siteHandler.AddSite)))
	mux.Handle("POST /api/sites/bulk", middleware.AuthMiddleware(cfg.JWTSecret)(http.HandlerFunc(siteHandler.BulkAddSites)))
	mux.Handle("GET /api/sites", middleware.AuthMiddleware(cfg.JWTSecret)(http.HandlerFunc(siteHandler.GetSites)))
	mux.Handle("DELETE /api/sites/", middleware.AuthMiddleware(cfg.JWTSecret)(http.HandlerFunc(siteHandler.DeleteSite)))
	mux.Handle("POST /api/sites/bulk-delete", middleware.AuthMiddleware(cfg.JWTSecret)(http.HandlerFunc(siteHandler.BulkDeleteSites)))
	mux.Handle("POST /api/sites/refresh", middleware.AuthMiddleware(cfg.JWTSecret)(http.HandlerFunc(siteHandler.RefreshSites)))
	mux.Handle("GET /api/verify-token", middleware.AuthMiddleware(cfg.JWTSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userHandler.VerifyToken(w, r, cfg.JWTSecret)
	})))
	mux.Handle("POST /api/telegram/link-code", middleware.AuthMiddleware(cfg.JWTSecret)(http.HandlerFunc(userHandler.GenerateTelegramLinkCode)))

	// Graceful shutdown
	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: mux,
	}

	go func() {
		log.Printf("Server starting on :%s...", cfg.ServerPort)
		log.Printf("Web UI available at: http://localhost:%s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server startup failed: %v", err)
		}
	}()

	// Ожидаем сигнал для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Printf("Shutting down server...")

	// Останавливаем checker и сервер
	cancel()
	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Printf("Server stopped")
}
