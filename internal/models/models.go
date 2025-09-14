package models

import (
	"time"
)

type User struct {
	ID             int       `json:"id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	PasswordHash   string    `json:"-"`
	TelegramChatID int64     `json:"telegram_chat_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type Site struct {
	ID          int       `json:"id"`
	URL         string    `json:"url"`
	UserID      int       `json:"user_id"`
	LastStatus  string    `json:"last_status"` // "UP", "DOWN", "UNKNOWN"
	LastChecked time.Time `json:"last_checked"`
	CreatedAt   time.Time `json:"created_at"`
}
