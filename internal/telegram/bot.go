package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aouxes/uptime-monitor/internal/storage"
)

type Bot struct {
	token   string
	client  *http.Client
	apiURL  string
	storage *storage.Storage
}

type Update struct {
	UpdateID int `json:"update_id"`
	Message  struct {
		MessageID int `json:"message_id"`
		From      struct {
			ID        int64  `json:"id"`
			Username  string `json:"username"`
			FirstName string `json:"first_name"`
		} `json:"from"`
		Chat struct {
			ID    int64  `json:"id"`
			Type  string `json:"type"`
			Title string `json:"title,omitempty"`
		} `json:"chat"`
		Text string `json:"text"`
		Date int64  `json:"date"`
	} `json:"message"`
}

type SendMessageRequest struct {
	ChatID    int64  `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

func NewBot(token string, storage *storage.Storage) *Bot {
	if token == "" {
		log.Printf("Warning: Telegram token is empty")
		return &Bot{
			token:   "",
			client:  &http.Client{Timeout: 30 * time.Second},
			apiURL:  "",
			storage: storage,
		}
	}

	// Проверяем формат токена
	if len(token) < 10 || !strings.Contains(token, ":") {
		log.Printf("Warning: Telegram token format seems invalid: %s...", token[:min(10, len(token))])
	}

	return &Bot{
		token:   token,
		client:  &http.Client{Timeout: 30 * time.Second},
		apiURL:  fmt.Sprintf("https://api.telegram.org/bot%s", token),
		storage: storage,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (b *Bot) testToken() error {
	resp, err := b.client.Get(b.apiURL + "/getMe")
	if err != nil {
		return fmt.Errorf("failed to test token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var response struct {
		OK          bool   `json:"ok"`
		ErrorCode   int    `json:"error_code,omitempty"`
		Description string `json:"description,omitempty"`
		Result      struct {
			ID        int64  `json:"id"`
			IsBot     bool   `json:"is_bot"`
			FirstName string `json:"first_name"`
			Username  string `json:"username"`
		} `json:"result,omitempty"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.OK {
		return fmt.Errorf("telegram API error %d: %s", response.ErrorCode, response.Description)
	}

	log.Printf("Bot info: @%s (%s) - ID: %d", response.Result.Username, response.Result.FirstName, response.Result.ID)
	return nil
}

func (b *Bot) Start(ctx context.Context) error {
	if b.token == "" {
		log.Printf("Telegram token not configured, bot not started")
		return nil
	}

	log.Printf("Starting Telegram bot with token: %s...", b.token[:10])

	// Проверяем токен, отправив тестовый запрос
	if err := b.testToken(); err != nil {
		log.Printf("Telegram token validation failed: %v", err)
		log.Printf("Bot will start in offline mode - notifications will be queued")
		// Не останавливаем бота, просто работаем в ограниченном режиме
	} else {
		log.Printf("Telegram token validated successfully")
	}

	offset := 0
	for {
		select {
		case <-ctx.Done():
			log.Printf("Telegram bot stopped")
			return nil
		default:
			updates, err := b.getUpdates(offset)
			if err != nil {
				log.Printf("Failed to get updates: %v", err)
				// Увеличиваем задержку при ошибках
				time.Sleep(10 * time.Second)
				continue
			}

			for _, update := range updates {
				if update.UpdateID >= offset {
					offset = update.UpdateID + 1
				}

				if err := b.handleUpdate(ctx, update); err != nil {
					log.Printf("Failed to handle update: %v", err)
				}
			}

			// Короткая задержка при успешном получении обновлений
			time.Sleep(1 * time.Second)
		}
	}
}

func (b *Bot) getUpdates(offset int) ([]Update, error) {
	url := fmt.Sprintf("%s/getUpdates?offset=%d&timeout=10", b.apiURL, offset)

	// Retry логика с экспоненциальной задержкой
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := b.client.Get(url)
		if err != nil {
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("failed to make request after %d attempts: %w", maxRetries, err)
			}
			log.Printf("Request attempt %d failed: %v, retrying...", attempt+1, err)
			time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
			continue
		}
		defer resp.Body.Close()

		var response struct {
			OK          bool     `json:"ok"`
			Result      []Update `json:"result"`
			ErrorCode   int      `json:"error_code,omitempty"`
			Description string   `json:"description,omitempty"`
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		if err := json.Unmarshal(body, &response); err != nil {
			log.Printf("Failed to decode response: %v", err)
			log.Printf("Response body: %s", string(body))
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		if !response.OK {
			log.Printf("Telegram API error: %d - %s", response.ErrorCode, response.Description)
			log.Printf("Response body: %s", string(body))
			return nil, fmt.Errorf("telegram API error %d: %s", response.ErrorCode, response.Description)
		}

		return response.Result, nil
	}

	return nil, fmt.Errorf("max retries exceeded")
}

func (b *Bot) handleUpdate(ctx context.Context, update Update) error {
	if update.Message.Text == "" {
		return nil
	}

	chatID := update.Message.Chat.ID
	text := update.Message.Text
	username := update.Message.From.Username

	log.Printf("Received message from %s (chat %d): %s", username, chatID, text)

	switch {
	case text == "/start":
		return b.sendMessage(chatID, "🤖 <b>Uptime Monitor Bot</b>\n\n"+
			"Этот бот поможет вам получать уведомления о статусе ваших сайтов.\n\n"+
			"<b>Доступные команды:</b>\n"+
			"/link <code>ваш_код</code> - Связать аккаунт\n"+
			"/unlink - Отвязать аккаунт\n"+
			"/status - Проверить статус связи\n"+
			"/help - Показать справку")

	case strings.HasPrefix(text, "/link "):
		code := strings.TrimSpace(strings.TrimPrefix(text, "/link "))
		return b.handleLinkCommand(ctx, chatID, code)

	case text == "/unlink":
		return b.handleUnlinkCommand(ctx, chatID)

	case text == "/status":
		return b.handleStatusCommand(ctx, chatID)

	case text == "/help":
		return b.sendMessage(chatID, "📖 <b>Справка по командам</b>\n\n"+
			"/link <code>код</code> - Связать ваш аккаунт с ботом\n"+
			"   Получите код в веб-интерфейсе в разделе настроек\n\n"+
			"/unlink - Отвязать аккаунт от бота\n\n"+
			"/status - Проверить, связан ли ваш аккаунт\n\n"+
			"/help - Показать эту справку")

	default:
		return b.sendMessage(chatID, "❓ Неизвестная команда. Используйте /help для справки.")
	}
}

func (b *Bot) handleLinkCommand(ctx context.Context, chatID int64, code string) error {
	if code == "" {
		return b.sendMessage(chatID, "❌ Пожалуйста, укажите код для связывания.\n"+
			"Использование: /link <code>ваш_код</code>")
	}

	// Ищем пользователя по коду связывания
	user, err := b.storage.GetUserByLinkCode(ctx, code)
	if err != nil {
		log.Printf("Failed to get user by link code: %v", err)
		return b.sendMessage(chatID, "❌ Ошибка при связывании аккаунта. Попробуйте позже.")
	}

	if user == nil {
		return b.sendMessage(chatID, "❌ Неверный код связывания. Проверьте код и попробуйте снова.")
	}

	// Обновляем chat ID пользователя
	if err := b.storage.UpdateUserTelegramChatID(ctx, user.ID, chatID); err != nil {
		log.Printf("Failed to update user telegram chat ID: %v", err)
		return b.sendMessage(chatID, "❌ Ошибка при связывании аккаунта. Попробуйте позже.")
	}

	// Удаляем использованный код
	if err := b.storage.DeleteLinkCode(ctx, code); err != nil {
		log.Printf("Failed to delete link code: %v", err)
	}

	return b.sendMessage(chatID, fmt.Sprintf("✅ <b>Аккаунт успешно связан!</b>\n\n"+
		"Пользователь: <code>%s</code>\n"+
		"Теперь вы будете получать уведомления о статусе ваших сайтов.", user.Username))
}

func (b *Bot) handleUnlinkCommand(ctx context.Context, chatID int64) error {
	// Находим пользователя по chat ID
	user, err := b.storage.GetUserByTelegramChatID(ctx, chatID)
	if err != nil {
		log.Printf("Failed to get user by telegram chat ID: %v", err)
		return b.sendMessage(chatID, "❌ Ошибка при отвязывании аккаунта. Попробуйте позже.")
	}

	if user == nil {
		return b.sendMessage(chatID, "❌ Ваш аккаунт не связан с ботом.")
	}

	// Отвязываем аккаунт
	if err := b.storage.UpdateUserTelegramChatID(ctx, user.ID, 0); err != nil {
		log.Printf("Failed to unlink user: %v", err)
		return b.sendMessage(chatID, "❌ Ошибка при отвязывании аккаунта. Попробуйте позже.")
	}

	return b.sendMessage(chatID, fmt.Sprintf("✅ <b>Аккаунт отвязан!</b>\n\n"+
		"Пользователь: <code>%s</code>\n"+
		"Вы больше не будете получать уведомления.", user.Username))
}

func (b *Bot) handleStatusCommand(ctx context.Context, chatID int64) error {
	// Находим пользователя по chat ID
	user, err := b.storage.GetUserByTelegramChatID(ctx, chatID)
	if err != nil {
		log.Printf("Failed to get user by telegram chat ID: %v", err)
		return b.sendMessage(chatID, "❌ Ошибка при проверке статуса. Попробуйте позже.")
	}

	if user == nil {
		return b.sendMessage(chatID, "❌ Ваш аккаунт не связан с ботом.\n\n"+
			"Для связывания используйте команду /link с кодом из веб-интерфейса.")
	}

	return b.sendMessage(chatID, fmt.Sprintf("✅ <b>Аккаунт связан!</b>\n\n"+
		"Пользователь: <code>%s</code>\n"+
		"Email: <code>%s</code>\n"+
		"Дата регистрации: %s\n\n"+
		"Вы получаете уведомления о статусе ваших сайтов.",
		user.Username, user.Email, user.CreatedAt.Format("02.01.2006")))
}

func (b *Bot) sendMessage(chatID int64, text string) error {
	message := SendMessageRequest{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "HTML",
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	resp, err := b.client.Post(b.apiURL+"/sendMessage", "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var response struct {
		OK          bool   `json:"ok"`
		ErrorCode   int    `json:"error_code,omitempty"`
		Description string `json:"description,omitempty"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("Failed to decode sendMessage response: %v", err)
		log.Printf("Response body: %s", string(body))
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.OK {
		log.Printf("SendMessage error: %d - %s", response.ErrorCode, response.Description)
		return fmt.Errorf("telegram API error %d: %s", response.ErrorCode, response.Description)
	}

	return nil
}
