package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Client struct {
	token  string
	client *http.Client
	apiURL string
}

type Message struct {
	ChatID    int64  `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

type SendMessageResponse struct {
	OK     bool `json:"ok"`
	Result struct {
		MessageID int `json:"message_id"`
	} `json:"result"`
}

func NewClient(token string) *Client {
	return &Client{
		token:  token,
		client: &http.Client{Timeout: 10 * time.Second},
		apiURL: fmt.Sprintf("https://api.telegram.org/bot%s", token),
	}
}

func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) error {
	if c.token == "" {
		log.Printf("Telegram token not configured, skipping notification")
		return nil
	}

	message := Message{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "HTML",
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.apiURL+"/sendMessage", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}

	var response SendMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.OK {
		return fmt.Errorf("telegram API returned error")
	}

	log.Printf("Telegram message sent successfully to chat %d", chatID)
	return nil
}

func (c *Client) SendSiteStatusNotification(ctx context.Context, chatID int64, siteURL, oldStatus, newStatus string) error {
	var emoji string
	var statusText string

	switch newStatus {
	case "UP":
		emoji = "✅"
		statusText = "ВЕРНУЛСЯ В СЕТЬ"
	case "DOWN":
		emoji = "❌"
		statusText = "НЕДОСТУПЕН"
	case "UNKNOWN":
		emoji = "❓"
		statusText = "НЕИЗВЕСТНО"
	default:
		emoji = "❓"
		statusText = newStatus
	}

	// Отправляем уведомление только при изменении статуса
	if oldStatus == newStatus {
		return nil
	}

	message := fmt.Sprintf(
		"%s <b>%s</b>\n\n"+
			"🌐 <b>Сайт:</b> %s\n"+
			"📊 <b>Статус:</b> %s\n"+
			"⏰ <b>Время:</b> %s",
		emoji,
		statusText,
		siteURL,
		newStatus,
		time.Now().Format("15:04:05 02.01.2006"),
	)

	return c.SendMessage(ctx, chatID, message)
}
