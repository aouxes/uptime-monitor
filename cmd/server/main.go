package main

import (
    "log"
    "net/http"

    "github.com/aouxes/uptime-monitor/internal/config"
    "github.com/aouxes/uptime-monitor/internal/storage"
)

func main() {
    cfg := config.Load()

    db, err := storage.New(cfg)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close() 

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("<h1>Uptime Monitor is running!</h1>"))
        w.Write([]byte("<p>Database connection: ✅ OK</p>"))
    })

    log.Printf("Server starting on :%s...", cfg.ServerPort)
    if err := http.ListenAndServe(":"+cfg.ServerPort, nil); err != nil {
        log.Fatalf("Server startup failed: %v", err)
    }
}