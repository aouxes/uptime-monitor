package handlers

import (
	"context"
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
