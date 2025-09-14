package checker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aouxes/uptime-monitor/internal/storage"
)

type Checker struct {
	storage    *storage.Storage
	interval   time.Duration
	workerPool *WorkerPool
}

func New(storage *storage.Storage, interval time.Duration, maxWorkers int) *Checker {
	return &Checker{
		storage:    storage,
		interval:   interval,
		workerPool: NewWorkerPool(storage, maxWorkers),
	}
}

// CheckAllSites использует worker pool
func (c *Checker) CheckAllSites(ctx context.Context) error {
	log.Printf("Starting check for all sites with %d workers...", c.workerPool.maxWorkers)

	sites, err := c.storage.GetAllSites(ctx)
	if err != nil {
		return fmt.Errorf("failed to get sites: %w", err)
	}

	log.Printf("Found %d sites for checking", len(sites))

	// Запускаем проверку асинхронно, не блокируя основной поток
	go func() {
		c.workerPool.ProcessSites(ctx, sites)
		log.Printf("Sites check completed")
	}()

	return nil
}

// Start запускает периодическую проверку
func (c *Checker) Start(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	if err := c.CheckAllSites(ctx); err != nil {
		log.Printf("Initial check failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.CheckAllSites(ctx); err != nil {
				log.Printf("Sites check failed: %v", err)
			}
		}
	}
}
