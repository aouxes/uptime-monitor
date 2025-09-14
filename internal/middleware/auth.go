package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/aouxes/uptime-monitor/internal/utils"
)

// contextKey — кастомный тип для ключей контекста (во избежание коллизий)
type contextKey string

const (
	UserIDKey contextKey = "user_id"
)

func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("AuthMiddleware: Processing request to %s", r.URL.Path)

			authHeader := r.Header.Get("Authorization")
			log.Printf("AuthMiddleware: Authorization header: %s", authHeader)

			if authHeader == "" {
				log.Printf("AuthMiddleware: No authorization header")
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Printf("AuthMiddleware: Invalid authorization format: %v", parts)
				http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]
			log.Printf("AuthMiddleware: Token: %s...", tokenString[:min(20, len(tokenString))])

			claims, err := utils.ParseJWT(tokenString, jwtSecret)
			if err != nil {
				log.Printf("AuthMiddleware: JWT validation failed: %v", err)
				log.Printf("AuthMiddleware: JWT Secret length: %d", len(jwtSecret))
				log.Printf("AuthMiddleware: Token length: %d", len(tokenString))
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			log.Printf("AuthMiddleware: User ID: %d", claims.UserID)
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
