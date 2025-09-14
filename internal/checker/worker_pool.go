package checker

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/aouxes/uptime-monitor/internal/models"
	"github.com/aouxes/uptime-monitor/internal/storage"
)

type WorkerPool struct {
	storage      *storage.Storage
	maxWorkers   int
	checkTimeout time.Duration
}

func NewWorkerPool(storage *storage.Storage, maxWorkers int) *WorkerPool {
	return &WorkerPool{
		storage:      storage,
		maxWorkers:   maxWorkers,
		checkTimeout: 15 * time.Second,
	}
}

func (wp *WorkerPool) ProcessSites(ctx context.Context, sites []models.Site) {
	var wg sync.WaitGroup
	sitesCh := make(chan models.Site, len(sites))

	// Запускаем worker'ов
	for i := 0; i < wp.maxWorkers; i++ {
		wg.Add(1)
		go wp.worker(ctx, &wg, sitesCh)
	}

	// Отправляем сайты в канал
	for _, site := range sites {
		sitesCh <- site
	}
	close(sitesCh)

	wg.Wait()
}

func (wp *WorkerPool) worker(ctx context.Context, wg *sync.WaitGroup, sitesCh <-chan models.Site) {
	defer wg.Done()

	for site := range sitesCh {
		select {
		case <-ctx.Done():
			return
		default:
			wp.processSite(ctx, site)
		}
	}
}

// processSite обрабатывает один сайт
func (wp *WorkerPool) processSite(ctx context.Context, site models.Site) {
	ctx, cancel := context.WithTimeout(ctx, wp.checkTimeout)
	defer cancel()

	// Вызываем статический метод CheckSite
	status, err := CheckSite(ctx, site.URL)
	if err != nil {
		log.Printf("❌ Failed to check site %s: %v", site.URL, err)
		status = "DOWN"
	}

	if err := wp.storage.UpdateSiteStatus(ctx, site.ID, status); err != nil {
		log.Printf("Failed to update site %s: %v", site.URL, err)
	}
}

func CheckSite(ctx context.Context, url string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	start := time.Now()
	resp, err := client.Head(url)
	checkTime := time.Since(start)

	if err != nil {
		log.Printf("❌ Site %s is DOWN (error: %v, time: %v)", url, err, checkTime)
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
