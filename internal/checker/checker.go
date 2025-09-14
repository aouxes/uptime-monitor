package checker

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aouxes/uptime-monitor/internal/storage"
)

type Checker struct {
	storage  *storage.Storage
	interval time.Duration
}

func New(storage *storage.Storage, interval time.Duration) *Checker {
	return &Checker{
		storage:  storage,
		interval: interval,
	}
}

func (c *Checker) CheckSite(ctx context.Context, url string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	start := time.Now()
	resp, err := client.Head(url)
	checkTime := time.Since(start)

	if err != nil {
		return "DOWN", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		log.Printf("✅ Site %s is UP (status: %d, time: %v)", url, resp.StatusCode, checkTime)
		return "UP", nil
	}

	log.Printf("❌ Site %s is DOWN (status: %d, time: %v)", url, resp.StatusCode, checkTime)
	return "DOWN", nil
}

func (c *Checker) CheckAllSites(ctx context.Context) error {
	log.Printf("Starting sites check...")

	sites, err := c.storage.GetAllSites(ctx)
	if err != nil {
		return fmt.Errorf("failed to get sites: %w", err)
	}

	log.Printf("Found %d sites to check", len(sites))

	for _, site := range sites {
		status, err := c.CheckSite(ctx, site.URL)
		if err != nil {
			log.Printf("Failed to check site %s: %v", site.URL, err)
			status = "DOWN"
		}

		if err := c.storage.UpdateSiteStatus(ctx, site.ID, status); err != nil {
			log.Printf("Failed to update site status %s: %v", site.URL, err)
		}

		time.Sleep(1 * time.Second)
	}

	log.Printf("Sites check completed")
	return nil
}

func (c *Checker) Start(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	if err := c.CheckAllSites(ctx); err != nil {
		log.Printf("Initial sites check failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Printf("Checker stopped")
			return
		case <-ticker.C:
			if err := c.CheckAllSites(ctx); err != nil {
				log.Printf("Sites check failed: %v", err)
			}
		}
	}
}
