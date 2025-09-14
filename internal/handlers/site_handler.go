package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/aouxes/uptime-monitor/internal/middleware"
	"github.com/aouxes/uptime-monitor/internal/models"
	"github.com/aouxes/uptime-monitor/internal/storage"
)

type SiteHandler struct {
	storage *storage.Storage
}

func NewSiteHandler(storage *storage.Storage) *SiteHandler {
	return &SiteHandler{storage: storage}
}

type AddSiteRequest struct {
	URL string `json:"url"`
}

func (h *SiteHandler) AddSite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем user_id из контекста (добавленного middleware)
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

	// TODO: Добавить валидацию URL

	site := &models.Site{
		URL:    req.URL,
		UserID: userID, // Связываем сайт с пользователем
	}

	// TODO: Реализовать метод CreateSite в storage
	log.Printf("Would create site: %+v for user_id: %d", site, userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Site added successfully (TODO: implement storage)",
		"url":     req.URL,
	})
}
