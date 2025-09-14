package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/aouxes/uptime-monitor/internal/models"
	"github.com/jackc/pgx/v5"
)

func (s *Storage) CreateUser(ctx context.Context, user *models.User) error {
	query := `
        INSERT INTO users (username, email, password_hash, telegram_chat_id)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at
    `

	err := s.db.QueryRow(ctx, query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.TelegramChatID,
	).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	log.Printf("User created successfully: ID=%d, Username=%s", user.ID, user.Username)
	return nil
}

func (s *Storage) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
        SELECT id, username, email, password_hash, telegram_chat_id, created_at
        FROM users 
        WHERE username = $1
    `

	var user models.User
	err := s.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.TelegramChatID,
		&user.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}
