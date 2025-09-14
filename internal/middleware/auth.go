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
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			claims, err := utils.ParseJWT(tokenString, jwtSecret)
			if err != nil {
				log.Printf("JWT validation failed: %v", err)
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
