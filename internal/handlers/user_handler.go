package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/aouxes/uptime-monitor/internal/models"
	"github.com/aouxes/uptime-monitor/internal/storage"
	"github.com/aouxes/uptime-monitor/internal/utils"
)

type UserHandler struct {
	storage *storage.Storage
}

func NewUserHandler(storage *storage.Storage) *UserHandler {
	return &UserHandler{storage: storage}
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if errors := utils.ValidateUser(req.Username, req.Email, req.Password); len(errors) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Validation failed",
			"details": errors,
		})
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Printf("Password hashing failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	ctx := context.Background()
	if err := h.storage.CreateUser(ctx, user); err != nil {
		log.Printf("Registration failed: %v", err)
		http.Error(w, "Registration failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User created successfully",
		"user_id": user.ID,
	})
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request, jwtSecret string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	user, err := h.storage.GetUserByUsername(ctx, req.Username)
	if err != nil || user == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateJWT(user, jwtSecret)
	if err != nil {
		log.Printf("JWT generation failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"token":   token,
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}
