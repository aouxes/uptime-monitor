package checker

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/aouxes/uptime-monitor/internal/models"
	"github.com/aouxes/uptime-monitor/internal/notifier"
	"github.com/aouxes/uptime-monitor/internal/storage"
)

type WorkerPool struct {
	storage      *storage.Storage
	maxWorkers   int
	checkTimeout time.Duration
	notifier     *notifier.Notifier
}

func NewWorkerPool(storage *storage.Storage, maxWorkers int, notifier *notifier.Notifier) *WorkerPool {
	return &WorkerPool{
		storage:      storage,
		maxWorkers:   maxWorkers,
		checkTimeout: 15 * time.Second,
		notifier:     notifier,
	}
}

func (wp *WorkerPool) ProcessSites(ctx context.Context, sites []models.Site) {
	var wg sync.WaitGroup
	sitesCh := make(chan models.Site, len(sites))

	log.Printf("Starting %d workers for parallel site checking...", wp.maxWorkers)

	// Запускаем worker'ов
	for i := 0; i < wp.maxWorkers; i++ {
		wg.Add(1)
		go wp.worker(ctx, &wg, sitesCh, i+1)
	}

	// Отправляем сайты в канал
	for _, site := range sites {
		sitesCh <- site
	}
	close(sitesCh)

	wg.Wait()
	log.Printf("All workers completed processing %d sites", len(sites))
}

func (wp *WorkerPool) worker(ctx context.Context, wg *sync.WaitGroup, sitesCh <-chan models.Site, workerID int) {
	defer wg.Done()

	for site := range sitesCh {
		select {
		case <-ctx.Done():
			return
		default:
			wp.processSite(ctx, site, workerID)
		}
	}
}

// processSite обрабатывает один сайт
func (wp *WorkerPool) processSite(ctx context.Context, site models.Site, workerID int) {
	ctx, cancel := context.WithTimeout(ctx, wp.checkTimeout)
	defer cancel()

	log.Printf("Worker %d: Checking site %s", workerID, site.URL)

	// Сохраняем старый статус для сравнения
	oldStatus := site.LastStatus

	// Вызываем статический метод CheckSite
	status, err := CheckSite(ctx, site.URL)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("Worker %d: Timeout checking site %s (%v)", workerID, site.URL, wp.checkTimeout)
		} else {
			log.Printf("Worker %d: Failed to check site %s: %v", workerID, site.URL, err)
		}
		status = "DOWN"
	} else {
		log.Printf("Worker %d: Site %s is %s", workerID, site.URL, status)
	}

	// Обновляем статус в базе данных
	if err := wp.storage.UpdateSiteStatus(ctx, site.ID, status); err != nil {
		log.Printf("Worker %d: Failed to update site %s status: %v", workerID, site.URL, err)
	} else {
		// Отправляем уведомление если статус изменился
		if oldStatus != status {
			if err := wp.notifier.NotifySiteStatusChange(ctx, site.ID, oldStatus, status); err != nil {
				log.Printf("Worker %d: Failed to send notification for site %s: %v", workerID, site.URL, err)
			}
		}
	}
}

func CheckSite(ctx context.Context, url string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return "DOWN", err
	}

	resp, err := client.Do(req)
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
