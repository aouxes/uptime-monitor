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

func (s *Storage) GetUserByID(ctx context.Context, userID int) (*models.User, error) {
	query := `
        SELECT id, username, email, password_hash, telegram_chat_id, created_at
        FROM users 
        WHERE id = $1
    `

	var user models.User
	err := s.db.QueryRow(ctx, query, userID).Scan(
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

func (s *Storage) GetUserByTelegramChatID(ctx context.Context, chatID int64) (*models.User, error) {
	query := `
        SELECT id, username, email, password_hash, telegram_chat_id, created_at
        FROM users 
        WHERE telegram_chat_id = $1
    `

	var user models.User
	err := s.db.QueryRow(ctx, query, chatID).Scan(
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
		return nil, fmt.Errorf("failed to get user by telegram chat ID: %w", err)
	}

	return &user, nil
}

func (s *Storage) UpdateUserTelegramChatID(ctx context.Context, userID int, chatID int64) error {
	query := `UPDATE users SET telegram_chat_id = $1 WHERE id = $2`

	_, err := s.db.Exec(ctx, query, chatID, userID)
	if err != nil {
		return fmt.Errorf("failed to update user telegram chat ID: %w", err)
	}

	log.Printf("Updated user %d telegram chat ID to %d", userID, chatID)
	return nil
}

func (s *Storage) GetUserByLinkCode(ctx context.Context, code string) (*models.User, error) {
	query := `
        SELECT u.id, u.username, u.email, u.password_hash, u.telegram_chat_id, u.created_at
        FROM users u
        JOIN link_codes lc ON u.id = lc.user_id
        WHERE lc.code = $1 AND lc.expires_at > NOW()
    `

	var user models.User
	err := s.db.QueryRow(ctx, query, code).Scan(
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
		return nil, fmt.Errorf("failed to get user by link code: %w", err)
	}

	return &user, nil
}

func (s *Storage) CreateLinkCode(ctx context.Context, userID int, code string) error {
	log.Printf("CreateLinkCode called: userID=%d, code=%s", userID, code)

	// Сначала удаляем старый код для этого пользователя (если есть)
	deleteQuery := `DELETE FROM link_codes WHERE user_id = $1`
	log.Printf("Deleting old codes for user %d", userID)
	_, err := s.db.Exec(ctx, deleteQuery, userID)
	if err != nil {
		log.Printf("Warning: failed to delete old codes: %v", err)
	}

	// Создаем новый код
	insertQuery := `
        INSERT INTO link_codes (user_id, code, expires_at)
        VALUES ($1, $2, NOW() + INTERVAL '10 minutes')
    `

	log.Printf("Executing insert query: %s", insertQuery)
	log.Printf("Query parameters: userID=%d, code=%s", userID, code)

	_, err = s.db.Exec(ctx, insertQuery, userID, code)
	if err != nil {
		log.Printf("Database error: %v", err)
		return fmt.Errorf("failed to create link code: %w", err)
	}

	log.Printf("Created link code for user %d", userID)
	return nil
}

func (s *Storage) DeleteLinkCode(ctx context.Context, code string) error {
	query := `DELETE FROM link_codes WHERE code = $1`

	_, err := s.db.Exec(ctx, query, code)
	if err != nil {
		return fmt.Errorf("failed to delete link code: %w", err)
	}

	log.Printf("Deleted link code: %s", code)
	return nil
}
