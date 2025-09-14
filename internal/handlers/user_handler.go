package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	"github.com/aouxes/uptime-monitor/internal/middleware"
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
	log.Printf("Login called")
	log.Printf("JWT Secret length: %d", len(jwtSecret))

	if r.Method != http.MethodPost {
		log.Printf("Invalid method: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Login attempt for user: %s", req.Username)

	ctx := context.Background()
	user, err := h.storage.GetUserByUsername(ctx, req.Username)
	if err != nil || user == nil {
		log.Printf("User not found: %s, error: %v", req.Username, err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	log.Printf("User found: ID=%d, Username=%s", user.ID, user.Username)

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		log.Printf("Invalid password for user: %s", req.Username)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	log.Printf("Password verified for user: %s", req.Username)

	token, err := utils.GenerateJWT(user, jwtSecret)
	if err != nil {
		log.Printf("JWT generation failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("JWT token generated successfully, length: %d", len(token))

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

func (h *UserHandler) VerifyToken(w http.ResponseWriter, r *http.Request, jwtSecret string) {
	log.Printf("VerifyToken called")

	if r.Method != http.MethodGet {
		log.Printf("Invalid method: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Middleware уже проверил токен, просто возвращаем успех
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		log.Printf("User ID not found in context")
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	log.Printf("Token verification successful for user ID: %d", userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":   true,
		"user_id": userID,
	})
}

// GenerateTelegramLinkCode генерирует код для связывания с Telegram
func (h *UserHandler) GenerateTelegramLinkCode(w http.ResponseWriter, r *http.Request) {
	log.Printf("GenerateTelegramLinkCode called")

	if r.Method != http.MethodPost {
		log.Printf("Invalid method: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		log.Printf("User not authenticated")
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	log.Printf("Generating link code for user ID: %d", userID)

	// Генерируем случайный код
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		log.Printf("Failed to generate random code: %v", err)
		http.Error(w, "Failed to generate code", http.StatusInternalServerError)
		return
	}
	code := hex.EncodeToString(bytes)
	log.Printf("Generated code: %s", code)

	// Сохраняем код в базе данных
	ctx := context.Background()
	if err := h.storage.CreateLinkCode(ctx, userID, code); err != nil {
		log.Printf("Failed to create link code: %v", err)
		http.Error(w, "Failed to create link code", http.StatusInternalServerError)
		return
	}

	log.Printf("Link code created successfully for user %d", userID)

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"code":       code,
		"expires_in": 600, // 10 минут в секундах
		"message":    "Код создан. Отправьте команду /link " + code + " боту в Telegram.",
	}

	log.Printf("Sending response: %+v", response)
	json.NewEncoder(w).Encode(response)
}
