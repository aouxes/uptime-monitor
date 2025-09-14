package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aouxes/uptime-monitor/internal/checker"
	"github.com/aouxes/uptime-monitor/internal/middleware"
	"github.com/aouxes/uptime-monitor/internal/models"
	"github.com/aouxes/uptime-monitor/internal/notifier"
	"github.com/aouxes/uptime-monitor/internal/storage"
)

type SiteHandler struct {
	storage  *storage.Storage
	notifier *notifier.Notifier
}

func NewSiteHandler(storage *storage.Storage, notifier *notifier.Notifier) *SiteHandler {
	return &SiteHandler{
		storage:  storage,
		notifier: notifier,
	}
}

type AddSiteRequest struct {
	URL string `json:"url"`
}

func (h *SiteHandler) AddSite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var req AddSiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.URL == "" || len(req.URL) < 10 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	site := &models.Site{
		URL:    req.URL,
		UserID: userID,
	}

	ctx := context.Background()
	if err := h.storage.CreateSite(ctx, site); err != nil {
		log.Printf("Failed to create site: %v", err)
		http.Error(w, "Failed to add site", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Site added successfully",
		"site_id": site.ID,
		"url":     site.URL,
	})
}
func (h *SiteHandler) GetSites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()
	sites, err := h.storage.GetUserSites(ctx, userID)
	if err != nil {
		log.Printf("Failed to get user sites: %v", err)
		http.Error(w, "Failed to get sites", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sites": sites,
		"count": len(sites),
	})
}

func (h *SiteHandler) DeleteSite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	siteID, err := getSiteIDFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid site ID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	if err := h.storage.DeleteSite(ctx, siteID, userID); err != nil {
		log.Printf("Failed to delete site: %v", err)
		http.Error(w, "Failed to delete site", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Site deleted successfully",
		"site_id": siteID,
	})
}

func getSiteIDFromRequest(r *http.Request) (int, error) {
	path := r.URL.Path
	idStr := path[len("/api/sites/"):]

	var siteID int
	_, err := fmt.Sscanf(idStr, "%d", &siteID)
	if err != nil {
		return 0, err
	}

	return siteID, nil
}

type BulkAddSitesRequest struct {
	URLs []string `json:"urls"`
}

func (h *SiteHandler) BulkAddSites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var req BulkAddSitesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req.URLs) == 0 {
		http.Error(w, "No URLs provided", http.StatusBadRequest)
		return
	}

	if len(req.URLs) > 50 {
		http.Error(w, "Too many URLs. Maximum 50 at once", http.StatusBadRequest)
		return
	}

	var results []map[string]interface{}
	ctx := context.Background()

	for _, url := range req.URLs {
		if url == "" {
			continue
		}

		if len(url) < 10 {
			results = append(results, map[string]interface{}{
				"url":     url,
				"status":  "error",
				"message": "Invalid URL",
			})
			continue
		}

		site := &models.Site{
			URL:    url,
			UserID: userID,
		}

		if err := h.storage.CreateSite(ctx, site); err != nil {
			results = append(results, map[string]interface{}{
				"url":     url,
				"status":  "error",
				"message": err.Error(),
			})
		} else {
			results = append(results, map[string]interface{}{
				"url":     url,
				"status":  "success",
				"site_id": site.ID,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Bulk add completed",
		"results": results,
		"total":   len(results),
		"success": countSuccess(results),
	})
}

func countSuccess(results []map[string]interface{}) int {
	count := 0
	for _, result := range results {
		if result["status"] == "success" {
			count++
		}
	}
	return count
}

type BulkDeleteSitesRequest struct {
	SiteIDs []int `json:"site_ids"`
}

func (h *SiteHandler) BulkDeleteSites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	var req BulkDeleteSitesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req.SiteIDs) == 0 {
		http.Error(w, "No site IDs provided", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	var results []map[string]interface{}
	var successCount int

	for _, siteID := range req.SiteIDs {
		if err := h.storage.DeleteSite(ctx, siteID, userID); err != nil {
			results = append(results, map[string]interface{}{
				"site_id": siteID,
				"status":  "error",
				"message": err.Error(),
			})
		} else {
			results = append(results, map[string]interface{}{
				"site_id": siteID,
				"status":  "success",
			})
			successCount++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Bulk delete completed",
		"results": results,
		"total":   len(req.SiteIDs),
		"success": successCount,
		"failed":  len(req.SiteIDs) - successCount,
	})
}

// RefreshSites обновляет статусы всех сайтов пользователя
func (h *SiteHandler) RefreshSites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()

	// Получаем все сайты пользователя
	sites, err := h.storage.GetUserSites(ctx, userID)
	if err != nil {
		log.Printf("Failed to get user sites for refresh: %v", err)
		http.Error(w, "Failed to get sites", http.StatusInternalServerError)
		return
	}

	if len(sites) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "No sites to refresh",
			"count":   0,
		})
		return
	}

	// Асинхронно обновляем статусы всех сайтов пользователя
	maxConcurrency := 10
	log.Printf("Starting manual refresh for user %d: %d sites (max concurrency: %d)", userID, len(sites), maxConcurrency)

	var wg sync.WaitGroup
	var updatedCount int64

	// Создаем канал для ограничения количества одновременных проверок
	semaphore := make(chan struct{}, maxConcurrency)

	for _, site := range sites {
		wg.Add(1)
		go func(s models.Site) {
			defer wg.Done()

			// Получаем разрешение на выполнение
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			log.Printf("Starting check for site %s", s.URL)

			// Создаем отдельный контекст с timeout для каждой проверки
			siteCtx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
			defer cancel()

			// Используем статический метод CheckSite для проверки
			status, err := checker.CheckSite(siteCtx, s.URL)
			if err != nil {
				if siteCtx.Err() == context.DeadlineExceeded {
					log.Printf("Timeout checking site %s (12s)", s.URL)
				} else {
					log.Printf("Failed to check site %s: %v", s.URL, err)
				}
				status = "DOWN"
			}

			// Получаем старый статус для сравнения
			oldSite, err := h.storage.GetSiteByID(ctx, s.ID)
			if err != nil {
				log.Printf("Failed to get old site status for %s: %v", s.URL, err)
			}

			// Обновляем статус в базе данных (используем оригинальный контекст для БД)
			if err := h.storage.UpdateSiteStatus(ctx, s.ID, status); err != nil {
				log.Printf("Failed to update site %s status: %v", s.URL, err)
			} else {
				log.Printf("Completed check for site %s: %s", s.URL, status)
				atomic.AddInt64(&updatedCount, 1)

				// Отправляем уведомление, если статус изменился
				if oldSite != nil && oldSite.LastStatus != status {
					log.Printf("Status changed for site %s: %s -> %s, sending notification", s.URL, oldSite.LastStatus, status)
					if h.notifier != nil {
						if err := h.notifier.NotifySiteStatusChange(ctx, s.ID, oldSite.LastStatus, status); err != nil {
							log.Printf("Failed to send notification for site %s: %v", s.URL, err)
						} else {
							log.Printf("Notification sent for site %s status change", s.URL)
						}
					} else {
						log.Printf("Notifier is nil, skipping notification for site %s", s.URL)
					}
				} else if oldSite != nil {
					log.Printf("Status unchanged for site %s: %s", s.URL, status)
				}
			}
		}(site)
	}

	// Ждем завершения всех горутин
	wg.Wait()

	log.Printf("Manual refresh completed: %d/%d sites updated", updatedCount, len(sites))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Sites refreshed successfully",
		"total":   len(sites),
		"updated": int(updatedCount),
	})
}
